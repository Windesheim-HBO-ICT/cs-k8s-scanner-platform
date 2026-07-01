# Scanner Platform testapp

Kleine testapp voor de Kubernetes-praktijkopdracht in `dt-ca`.

De app is bedoeld om zichtbaar te maken of een cluster goed omgaat met Deployment, Service, ConfigMap, Secret, rolling updates en scheduling over meerdere nodes.

## Endpoints

| Endpoint | Doel |
| --- | --- |
| `GET /healthz` | Basis-healthcheck. |
| `GET /readyz` | Controleert verplichte environment variables. |
| `GET /info` | Toont versie, pod, namespace, node en configuratiecontext. |
| `GET /config-check` | Toont waarden uit ConfigMap. |
| `GET /secret-check` | Toont of secrets aanwezig zijn, gemaskeerd. |
| `POST /scan-test` | Accepteert een testpakketcode en toont welke pod de request verwerkte. |

## Verplichte environment variables

Uit ConfigMap:

- `SP_FEATURE_FLAG`
- `SP_REGION`
- `SP_SCAN_MODE`

Uit Secret:

- `SP_API_TOKEN`
- `SP_DB_PASSWORD`

Runtime metadata:

- `APP_VERSION`
- `POD_NAME`
- `POD_NAMESPACE`
- `NODE_NAME`

## Lokaal draaien

```bash
docker run --rm -p 8080:8080 \
  -e SP_FEATURE_FLAG=true \
  -e SP_REGION=groningen \
  -e SP_SCAN_MODE=demo \
  -e SP_API_TOKEN=test-token \
  -e SP_DB_PASSWORD=test-password \
  -e APP_VERSION=v1.0.0 \
  -e POD_NAME=local \
  -e POD_NAMESPACE=scanner-platform \
  -e NODE_NAME=local-node \
  ghcr.io/windesheim-hbo-ict/cs-k8s-scanner-platform:v1.0.0
```

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/readyz
curl -s -X POST http://localhost:8080/scan-test \
  -H 'content-type: application/json' \
  -d '{"package_code":"DPD-TEST-001"}'
```

## Image bouwen en publiceren

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg APP_VERSION=v1.0.0 \
  -t ghcr.io/windesheim-hbo-ict/cs-k8s-scanner-platform:v1.0.0 \
  --push .
```
