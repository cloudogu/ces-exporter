FROM alpine:3.21.2

RUN apk update && apk upgrade && apk --no-cache add bash openssh rsync \
    && addgroup -S -g 1000 ces-exporter \
    && adduser -S -h /home/ces-exporter -s /bin/bash -G ces-exporter -u 1000 ces-exporter \
    && mkdir -p /home/ces-exporter/.ssh /home/ces-exporter/.ssh/ssh_host_keys \
    && chown -R ces-exporter:ces-exporter /home/ces-exporter/.ssh \
    && chmod 700 /home/ces-exporter/.ssh

# Generate SSH host keys as root
RUN ssh-keygen -t rsa -f /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_rsa_key -N "" && \
    ssh-keygen -t ecdsa -f /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_ecdsa_key -N "" && \
    ssh-keygen -t ed25519 -f /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_ed25519_key -N "" && \
    chown -R ces-exporter:ces-exporter /home/ces-exporter/.ssh/ssh_host_keys &&  \
    chmod 600 /home/ces-exporter/.ssh/ssh_host_keys/*


# Allow non-root user to run SSH
# RUN echo "ces-exporter ALL=(ALL) NOPASSWD: /usr/sbin/sshd" > /etc/sudoers.d/ces-exporter

# Configure SSHD to allow non-root user login
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' /etc/ssh/sshd_config && \
    echo "AllowUsers ces-exporter" >> /etc/ssh/sshd_config && \
    echo "HostKey /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_rsa_key" >> /etc/ssh/sshd_config && \
    echo "HostKey /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_ecdsa_key" >> /etc/ssh/sshd_config && \
    echo "HostKey /home/ces-exporter/.ssh/ssh_host_keys/ssh_host_ed25519_key" >> /etc/ssh/sshd_config

# Setup SSH
# https://docs.docker.com/engine/examples/running_ssh_service/
EXPOSE 22

USER ces-exporter

COPY ./resources /

CMD ["/startup.sh"]