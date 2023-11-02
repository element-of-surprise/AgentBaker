function Start-InstallGPUDriver {
    param(
        [Parameter(Mandatory = $false)]
        [bool]$EnableInstall
    )
  
    if (-not $EnableInstall) {
        Write-Log "ConfigGPUDriverIfNeeded is false. GPU driver installation skipped as per configuration."
        return $false
    }
  
    Write-Log "ConfigGPUDriverIfNeeded is true. GPU driver installation started as per configuration."
  
    $RebootNeeded = $false
  
    if (-not $PSScriptRoot) {
        $PSScriptRoot = Split-Path $MyInvocation.InvocationName
    }
    if ( -not $env:Path.Contains( "$PSScriptRoot;") ) {
        $env:Path = "$PSScriptRoot;$env:Path"
    }
  
    try {
        $FatalError = @()
  
        $LogFolder = "$PSScriptRoot\.."
  
        $Reboot = @{
            Needed = $false
        }
  
        Write-Log "Attempting to install Nvidia driver..."
  
        # Get the SetupTarget based on the input
        $Setup = Get-Setup
        $SetupTarget = $Setup.Target
        $Reboot.Needed = $Setup.RebootNeeded
        Write-Log "Setup complete"
  
        Add-DriverCertificate $Setup.CertificateUrl
        Write-Log "Certificate in store"
  
        Write-Log "Installing $SetupTarget ..."
        try {
            $InstallLogFolder = "$LogFolder\NvidiaInstallLog"
            $Arguments = "-s -n -log:$InstallLogFolder -loglevel:6"
      
            $p = Start-Process -FilePath $SetupTarget -ArgumentList $Arguments -PassThru
  
            $Timeout = 10 * 60 # in seconds. 10 minutes for timeout of the installation
            Wait-Process -InputObject $p -Timeout $Timeout -ErrorAction Stop
      
            # check if installation was successful
            if ($p.ExitCode -eq 0 -or $p.ExitCode -eq 1) {
                # 1 is issued when reboot is required after success
                Write-Log "GPU Driver Installation Success. Code: $($p.ExitCode)"
            }
            else {
                $ErrorMsg = "GPU Driver Installation Failed! Code: $($p.ExitCode)"
                Write-Log $ErrorMsg
                Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_FAILED -ErrorMessage $ErrorMsg
            }
  
            if ($Reboot.Needed -or $p.ExitCode -eq 1) {
                Write-Log "Reboot is needed for this GPU Driver..."
                $RebootNeeded = $true
            }
            return $RebootNeeded
        }
        catch [System.TimeoutException] {
            $ErrorMsg = "Timeout $Timeout s exceeded. Stopping the installation process. Reboot for another attempt."
            Write-Log $ErrorMsg
            Stop-Process -InputObject $p -Force
            Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_TIMEOUT -ErrorMessage $ErrorMsg
        }
        catch {
            $ErrorMsg = "Exception: $($_.ToString())"
            Write-Log $ErrorMsg # the status file may get over-written when the agent re-attempts this step
            Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_FAILED -ErrorMessage $ErrorMsg
        }
      
    }
    catch {
        $FatalError += $_
        $errorCount = $FatalError.Count
        $ErrorMsg = "A fatal error occurred. Number of errors: $errorCount. Error details: $($_ | Out-String)"
        Write-Log $ErrorMsg
        Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_FAILED -ErrorMessage $ErrorMsg
    }
}
  
function Get-Setup {
    [OutputType([hashtable])]
      
    # Choose driver and specific properties
    $Driver = Select-Driver
      
    if ($Driver.Url -eq $null -or $Driver.CertificateUrl -eq $null) {
        $ErrorMsg = "DriverURL or DriverCertificateURL are not properly specified."
        Write-Log $ErrorMsg
        Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_URL_NOT_SET -ErrorMessage $ErrorMsg
    }
      
    $Setup = @{
        RebootNeeded   = $Driver.RebootNeeded
        CertificateUrl = $Driver.CertificateUrl
    }
          
    $InstallFolder = "$PSScriptRoot\.."
      
    $source = $Driver.Url
    $dest = "$InstallFolder\$($Driver.InstallExe)"
  
    Get-DriverFile $source $dest
  
    $Setup.Target = "$InstallFolder\$($Driver.InstallExe)"
  
    return $Setup
}
  
function Select-Driver {
    [OutputType([hashtable])]
  
    # Set some default values
    $Driver = @{
        # Hard coding the FwLink for resources.json
        ResourceUrl  = "https://go.microsoft.com/fwlink/?linkid=2181101"
        ResourceFile = "$PSScriptRoot\..\resources.json"
        InstallExe   = "install.exe"
        SetupExe     = "setup.exe"
        RebootNeeded = $false
    }
    $Index = @{
        OS = 0 # Windows
    }
  
    # Get OS
    # Using the following instead of VM vmetadata ( $Compute.sku -match ('.*2016-Datacenter.*|.*RS.*Pro.*') ) # 2016, Win 10
    # to accomodate various image publishers and their SKUs
    $OSMajorVersion = (Get-CimInstance Win32_OperatingSystem).version.split('.')[0]
    if ( $OSMajorVersion -eq 10 ) {
        # Win 2016, Win 10
        $Index.OSVersion = 0
    }
    else {
        # ( $OSMajorVersion -eq 6 ) Win 2012 R2
        $Index.OSVersion = 1
    }  
  
    try {
        $Compute = Get-VmData
        $vmSize = $Compute.vmSize
    }
    catch {
        $ErrorMsg = "Failed to query the SKU information."
        Write-Log $ErrorMsg
        Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_SKU_INFO_NOT_FOUND -ErrorMessage $ErrorMsg
    }
  
    # Not an AzureStack scenario
    if ($vmSize -ne $null) {
        if ( ($Compute.vmSize -Match "_NC") -or
          ($Compute.vmSize -Match "_ND") ) {
            # Override if GRID driver is desired on ND or NC VMs
            $Index.Driver = 0 # CUDA
        }
        elseif ( $Compute.vmSize -Match "_NV" ) {
            # NV or Grid on ND or Grid on NC
            $Index.Driver = 1 # GRID
            $Driver.RebootNeeded = $true
        }
        else {
            $ErrorMsg = "VM type $($Compute.vmSize) is not an N-series VM. Not attempting to deploy extension! More information on troubleshooting is available at https://aka.ms/NvidiaGpuDriverWindowsExtensionTroubleshoot"
            Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_VM_SIZE_NOT_SUPPORTED -ErrorMessage $ErrorMsg
        }
    }
  
  
    # Download and read the resources table
    Get-DriverFile $Driver.ResourceUrl $Driver.ResourceFile
    $Resources = (Get-Content $Driver.ResourceFile) -join "`n" | ConvertFrom-JSON
  
    # Get the certificate url
    $Driver.CertificateUrl = $Resources.OS[$Index.OS].Certificate.FwLink
  
    # Get the driver version
    $Driver.ObjectArray = $Resources.OS[$Index.OS].Version[$Index.OSVersion].Driver[$Index.Driver].Version
    $Driver.Object = $Driver.ObjectArray[0] # latest driver
  
    $Driver.Version = $Driver.Object.Num
    $Driver.Url = $Driver.Object.FwLink
  
  
    # $Driver.SetupFolder is set based on OS and Driver Type
    # This cannot be made standard currently as there doesn't seem to be a way to pass the extraction/setup folder in silent mode
    if ($Index.Driver -eq 0) {
        # CUDA
        if ($Index.OSVersion -eq 0) {
            # Win 2016, Win 10
            $Driver.SetupFolder = "C:\NVIDIA\DisplayDriver\$($Driver.Version)\Win10_64\International"
        }
        else {
            # Win 2012 R2
            $Driver.SetupFolder = "C:\NVIDIA\DisplayDriver\$($Driver.Version)\Win8_Win7_64\International"
        }
    }
    else {
        # GRID
        $Driver.SetupFolder = "C:\NVIDIA\$($Driver.Version)"
    }
  
  
    return $Driver
}
  
function Get-VmData {
    [OutputType([hashtable])]
  
    # Retry to overcome failure in retrieving VM metadata
    $Loop = $true
    $RetryCount = 0
    $RetryCountMax = 10
    do {
        try {
            $Compute = Invoke-RestMethod -Headers @{"Metadata" = "true" } -URI http://169.254.169.254/metadata/instance/compute?api-version=2017-08-01 -Method get
            $Loop = $false
        }
        catch {
            if ($RetryCount -gt $RetryCountMax) {
                $Loop = $false
                $ErrorMsg = "Failed to retrieve VM metadata after $RetryCountMax attempts. Exiting! More information on troubleshooting is available at https://aka.ms/NvidiaGpuDriverWindowsExtensionTroubleshoot"
                Write-Log $ErrorMsg
                Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_SKU_INFO_NOT_FOUND -ErrorMessage $ErrorMsg
            }
            else {
                Start-Sleep -Seconds ($RetryCount * 2 + 1)
                $RetryCount++
            }
        }
    } while ($Loop)
  
    return $Compute
}
  
function Get-DriverFile {
    param(
        [Parameter(Mandatory = $True)]
        [ValidateNotNullOrEmpty()]
        [string] $source,
    
        [Parameter(Mandatory = $True)]
        [ValidateNotNullOrEmpty()]
        [string] $dest
    )
    
    function GetUsingWebClient {
        # Store Security Protocols
        $protocols = [Net.ServicePointManager]::SecurityProtocol
        # Add Tls12 to Security Protocols
        [Net.ServicePointManager]::SecurityProtocol = ([Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12)
  
        $wc = New-Object System.Net.WebClient
        $start_time = Get-Date
        $wc.DownloadFile($source, $dest)
        Write-Log "Time taken: $((Get-Date).Subtract($start_time).Seconds) second(s)"
  
        # Reset Security Protocols
        [Net.ServicePointManager]::SecurityProtocol = $protocols
    }
  
    Write-Log "Downloading from $source to $dest"
  
    # Retry to overcome failure in downloading
    $Loop = $true
    $RetryCount = 0
    $RetryCountMax = 10
    do {
        try {
            GetUsingWebClient
            $Loop = $false
        }
        catch {
            if ($RetryCount -gt $RetryCountMax) {
                $Loop = $false
                $ErrorMsg = "Failed to download $source after $RetryCountMax attempts. Exiting! More information on troubleshooting is available at https://aka.ms/NvidiaGpuDriverWindowsExtensionTroubleshoot"
                Write-Log $ErrorMsg
                Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_GPU_DRIVER_INSTALLATION_DOWNLOAD_FAILURE -ErrorMessage $ErrorMsg
            }
            else {
                Start-Sleep -Seconds ($RetryCount * 2 + 1)
                $RetryCount++
            }
        }
    } while ($Loop)
}
  
function Add-DriverCertificate {
    param(
        [Parameter(Mandatory = $True)]
        [ValidateNotNullOrEmpty()]
        [string] $link
    )
  
    $Cert = Get-DriverCertificate $link
  
    Write-Log 'Adding Certificate ...'
    $Store = Get-Item $Cert.Store
    $Store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
    $ImportedCert = Import-Certificate -Filepath $Cert.File -CertStoreLocation $Cert.Store
    $CertInStore = $Store.Certificates | Where-Object { $_.Thumbprint -eq $ImportedCert.Thumbprint }
  
    if ($CertInStore) {
        # Perform additional validation checks on the certificate
        if ($CertInStore.NotAfter -gt (Get-Date)) {
            Write-Log 'Certificate is valid and has not expired.'
        }
        else {
            Write-Log 'Certificate has expired.'
        }
  
        # Check if the certificate is trusted
        if ($CertInStore.Verify()) {
            Write-Log 'Certificate is trusted.'
        }
        else {
            Write-Log 'Certificate is not trusted.'
        }
    }
    else {
        Write-Log 'Certificate is not valid.'
    }
    Write-Log 'Certificate added.'
    $Store.Close()
}
   
function Get-DriverCertificate {
    param(
        [Parameter(Mandatory = $True)]
        [ValidateNotNullOrEmpty()]
        [string] $link
    )
  
    # Defaults
    $Cert = @{
        File      = "$PSScriptRoot\..\nvidia.cer"
        Store     = "Cert:\LocalMachine\TrustedPublisher\"
        NewObject = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2
    }
  
    # Download the .cer file
    $source = $link
    $dest = $Cert.File
    Get-DriverFile $source $dest
  
    # New Certificate object to check for properties
    $Cert.NewObject.Import($Cert.File)
  
    return $Cert
}