[Unit]
Description=Kubelet
ConditionPathExists=/usr/local/bin/kubelet
Wants=network-online.target containerd.service
After=network-online.target containerd.service

[Service]
Restart=always
RestartSec=2
TimeoutStartSec=270 
EnvironmentFile=/etc/default/kubelet
SuccessExitStatus=143

ExecStart=/usr/local/bin/kubelet \
        --enable-server \
        --node-labels="${KUBELET_NODE_LABELS}" \
        --v=2 \
        --volume-plugin-dir=/etc/kubernetes/volumeplugins \
        $KUBELET_TLS_BOOTSTRAP_FLAGS \
        $KUBELET_CONFIG_FILE_FLAGS \
        $KUBELET_CONTAINERD_FLAGS \
        $KUBELET_CONTAINER_RUNTIME_FLAG \
        $KUBELET_CGROUP_FLAGS \
        $KUBELET_FLAGS

[Install]
WantedBy=multi-user.target
