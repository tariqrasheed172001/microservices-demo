FROM grafana/grafana:latest

# Optional: configure default credentials
ENV GF_SECURITY_ADMIN_USER=admin
ENV GF_SECURITY_ADMIN_PASSWORD=admin

EXPOSE 3000
