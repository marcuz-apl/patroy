# Patroy Docker & Cloud Deployment

Patroy provides an ultra-lightweight multi-stage Docker image (~60MB footprint) packaged with headless Chromium, NSS, FreeType, HarfBuzz, and CA-certificates.

---

## 1. Quickstart with Docker Compose

By default, Patroy runs on port **`4023`**.

```bash
# Start container in detached mode
docker compose up -d

# Check service logs
docker compose logs -f

# Test service health endpoint
curl http://localhost:4023/health

# Stop container
docker compose down
```

---

## 2. Building and Running Directly via Docker

### Build Image
```bash
docker build -t patroy:latest .
```

### Run Container
```bash
docker run -d \
  --name patroy-service \
  --restart unless-stopped \
  --security-opt seccomp=unconfined \
  --shm-size=1gb \
  -p 4023:4023 \
  patroy:latest
```

> [!TIP]
> **Shared Memory Flag**: When running high-concurrency Chromium in Docker, `--shm-size=1gb` prevents browser crashes caused by default small `/dev/shm` buffers.

---

## 3. Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: patroy
  labels:
    app: patroy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: patroy
  template:
    metadata:
      labels:
        app: patroy
    spec:
      containers:
        - name: patroy
          image: ghcr.io/marcuz-apl/patroy:v0.4.0
          ports:
            - containerPort: 4023
          resources:
            requests:
              cpu: "250m"
              memory: "256Mi"
            limits:
              cpu: "2000m"
              memory: "1024Mi"
          livenessProbe:
            httpGet:
              path: /health
              port: 4023
            initialDelaySeconds: 5
            periodSeconds: 10
```
