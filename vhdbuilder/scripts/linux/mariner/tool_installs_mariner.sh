#!/bin/bash

echo "Sourcing tool_installs_mariner.sh"

installAscBaseline() {
   echo "Mariner TODO: installAscBaseline"
}

installBcc() {
    echo "Installing BCC tools..."
    dnf_makecache || exit $ERR_APT_UPDATE_TIMEOUT
    dnf_install 120 5 25 bcc-tools || exit $ERR_BCC_INSTALL_TIMEOUT
    echo "Installing BCC examples..."
    dnf_install 120 5 25 bcc-examples || exit $ERR_BCC_INSTALL_TIMEOUT
}

addMarinerNvidiaRepo() {
    if [[ $OS_VERSION == "2.0" ]]; then 
        MARINER_NVIDIA_REPO_FILEPATH="/etc/yum.repos.d/mariner-nvidia.repo"
        touch "${MARINER_NVIDIA_REPO_FILEPATH}"
        cat << EOF > "${MARINER_NVIDIA_REPO_FILEPATH}"
[mariner-official-nvidia]
name=CBL-Mariner Official Nvidia 2.0 x86_64
baseurl=https://packages.microsoft.com/cbl-mariner/2.0/prod/nvidia/x86_64
gpgkey=file:///etc/pki/rpm-gpg/MICROSOFT-RPM-GPG-KEY file:///etc/pki/rpm-gpg/MICROSOFT-METADATA-GPG-KEY
gpgcheck=1
repo_gpgcheck=1
enabled=1
skip_if_unavailable=True
sslverify=1
EOF
    fi
}

forceEnableIpForward() {
    CONFIG_FILEPATH="/etc/sysctl.d/99-force-bridge-forward.conf"
    touch ${CONFIG_FILEPATH}
    cat << EOF >> ${CONFIG_FILEPATH}
    net.ipv4.ip_forward = 1
    net.ipv4.conf.all.forwarding = 1
    net.ipv6.conf.all.forwarding = 1
    net.bridge.bridge-nf-call-iptables = 1
EOF
}

# The default 99-dhcp-en config on Mariner attempts to assign an IP address
# to the eth1 virtual function device, which delays cluster setup by 2 minutes.
# This workaround makes it so that dhcp is only enabled on eth0.
setMarinerNetworkdConfig() {
    CONFIG_FILEPATH="/etc/systemd/network/99-dhcp-en.network"
    touch ${CONFIG_FILEPATH}
    cat << EOF > ${CONFIG_FILEPATH} 
    [Match]
    Name=eth0

    [Network]
    DHCP=yes
    IPv6AcceptRA=no
EOF
# On Mariner 2.0 Marketplace images, the default systemd network config
# has an additional change that prevents Mariner from changing IP addresses
# every reboot
if [[ $OS_VERSION == "2.0" ]]; then 
    cat << EOF >> ${CONFIG_FILEPATH}

    [DHCPv4]
    SendRelease=false
EOF
fi
}


listInstalledPackages() {
    rpm -qa
}

# disable and mask all UU timers/services
disableDNFAutomatic() {
    # Make sure dnf-automatic is running with the notify timer rather than the auto install timer
    systemctlEnableAndStart dnf-automatic-notifyonly.timer || exit $ERR_SYSTEMCTL_START_FAIL

    # Ensure the automatic install timer is disabled. 
    # systemctlDisableAndStop adds .service to the end which doesn't work on timers.
    systemctl disable dnf-automatic-install.service || exit 1
    systemctl mask dnf-automatic-install.service || exit 1

    systemctl stop dnf-automatic-install.timer || exit 1
    systemctl disable dnf-automatic-install.timer || exit 1
    systemctl mask dnf-automatic-install.timer || exit 1
}

# Regardless of UU mode, ensure check-restart is running
enableCheckRestart() {
  # Even if UU is disabled, we should still run check-restart so that kured
  # will work as expected if it is installed.
  # At 8:000:00 UTC check if a reboot-required package was installed
  # Touch /var/run/reboot-required if a reboot required package was installed.
  # This helps avoid a Mariner specific reboot check command in kured.
  systemctlEnableAndStart check-restart.timer || exit $ERR_SYSTEMCTL_START_FAIL
}

# There are several issues in default file permissions when trying to run AMA and ASA extensions.
# These will be resolved in an upcoming base image. Work around them here for now.
fixCBLMarinerPermissions() {
# Set the dmi/id/product_uuid permissions to 444 so that mdsd can read the VM unique ID
# https://github.com/Azure/WALinuxAgent/blob/develop/config/99-azure-product-uuid.rules
# Future base images will include this config in the WALinuxAgent package.
    CONFIG_FILEPATH="/etc/udev/rules.d/99-azure-product-uuid.rules"
    touch ${CONFIG_FILEPATH}
    cat << EOF > ${CONFIG_FILEPATH}
SUBSYSTEM!="dmi", GOTO="product_uuid-exit"
ATTR{sys_vendor}!="Microsoft Corporation", GOTO="product_uuid-exit"
ATTR{product_name}!="Virtual Machine", GOTO="product_uuid-exit"
TEST!="/sys/devices/virtual/dmi/id/product_uuid", GOTO="product_uuid-exit"

RUN+="/bin/chmod 0444 /sys/devices/virtual/dmi/id/product_uuid"

LABEL="product_uuid-exit"

EOF

# /etc/rsyslog.d is 750 but should be 755 so non root users can read the configs
# This occurs because the umask in Mariner is 0027 and packer_source.sh created the folder
# Future base images will already have rsyslog installed with 755 /etc/rsyslog.d
    chmod 755 /etc/rsyslog.d
}

enableMarinerKata() {
    # Enable the mshv boot path
    sudo sed -i -e 's@menuentry "CBL-Mariner"@menuentry "Dom0" {\n    search --no-floppy --set=root --file /EFI/Microsoft/Boot/bootmgfw.efi\n        chainloader /EFI/Microsoft/Boot/bootmgfw.efi\n}\n\nmenuentry "CBL-Mariner"@'  /boot/grub2/grub.cfg

    # kata-osbuilder-generate is responsible for triggering the kata-osbuilder.sh script, which uses
    # dracut to generate an initrd for the nested VM using binaries from the Mariner host OS.
    systemctlEnableAndStart kata-osbuilder-generate
}

activateNfConntrack() {
    # explicitly activate nf_conntrack module so associated sysctls can be properly set 
    echo nf_conntrack >> /etc/modules-load.d/contrack.conf
}

installFIPS() {

    echo "Installing FIPS..."

    # Install necessary rpm pacakages
    dnf_install 120 5 25 grubby || exit $ERR_APT_INSTALL_TIMEOUT
    dnf_install 120 5 25 dracut-fips || exit $ERR_APT_INSTALL_TIMEOUT

    # Add the boot= cmd line parameter if the boot dir is not the same as the root dir
    boot_dev="$(df /boot/ | tail -1 | cut -d' ' -f1)"
    root_dev="$(df / | tail -1 | cut -d' ' -f1)"
    if [ ! "$root_dev" == "$boot_dev" ]; then
        boot_uuid="UUID=$(blkid $boot_dev -s UUID -o value)"

        # Enable FIPS mode and modify boot directory
        if grub2-editenv - list | grep -q kernelopts;then
                grub2-editenv - set "$(grub2-editenv - list | grep kernelopts) fips=1 boot=$boot_uuid"
        else
                grubby --update-kernel=ALL --args="fips=1 boot=$boot_uuid"
        fi
    fi
    
}

installFedRAMP() {

    echo "Configuring openssl default cipher suites..."

    sed -i '/= new_oids/a openssl_conf            = default_conf' /etc/pki/tls/openssl.cnf
    cat << EOF >> test.sh

    [default_conf] 

    ssl_conf = ssl_sect
    
    [ssl_sect]
    
    system_default = system_default_sect 

    [system_default_sect] 

    MinProtocol = TLSv1.2 

    CipherString = ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256 
    EOF



<<comments

    script_dir="$(dirname "$(realpath "$0")")"

    mkdir "$script_dir/apply_logs"
    echo "Apply scripts" 1>&2
    touch "$script_dir/fail.txt"
    touch "$script_dir/success.txt"
    touch $script_dir/failure_details.txt
    
    for script in $(find "$script_dir/rhel8" -name '*.sh' | sort -u); do
        scriptname="$(basename "${script}")"
        prunedname=$(echo "${scriptname#*-}" | cut -d'.' -f1)

        echo "checking '${prunedname}'"  1>&2
        if grep -q -E "^${prunedname}\$" "$script_dir/skip_list.txt" ; then
            # If we are running live scripts, run those anyways
            if [[ "${run_live}" == "yes" ]]; then
                if ! grep -q -E "^${prunedname}\$" "$script_dir/live_machine_only.txt" ; then
                    echo "Skipping ${script} since its in skip_list.txt but not also in live_machine_only.txt" 1>&2
                    continue
                fi
            else
                echo "Skipping ${script} since its in skip_list.txt" 1>&2
                continue
            fi
        fi

        if [[ "${marketplace}" == "yes" ]]; then
        if grep -q -E "^${prunedname}\$" "$script_dir/marketplace_skip_list.txt" ; then
            echo "Skipping ${script} since its in marketplace_skip_list.txt" 1>&2
            continue
        fi
        fi

        if [[ "${skip_apply}" == "yes" ]]; then
            echo "Skipping ${script} due to --skip_apply"  1>&2
            echo "Skipped ${script}" > "$script_dir/apply_logs/$(basename "${script}").log"
        else
            echo "Running ${script}" 1>&2
            out=$(${script} 2>&1)
            res=$?
            if [[ ${res} -ne 0 ]]; then
                basename "${script}" >> "$script_dir/fail.txt"
            else
                basename "${script}" >> "$script_dir/success.txt"
            fi
            echo "$out" > "$script_dir/apply_logs/$(basename "${script}").log"
        fi
    done

comments

}
