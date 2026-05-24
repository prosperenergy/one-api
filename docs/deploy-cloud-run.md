Cloud Run deploy (CI/CD) for one-api

This document explains how to deploy one-api to Google Cloud Run using GitHub Actions and how to run locally using Colima (mac).

Required GitHub repository secrets
- GCP_PROJECT: GCP project id
- GCP_SA_KEY: Service account JSON key. Minimum roles: Cloud Run Admin, Service Account User, Storage Admin (or Artifact Registry Admin) and optionally Artifact Registry Writer.
- CLOUD_RUN_REGION: e.g. us-central1
- ONEAPI_JWT_SECRET: random secret for one-api session tokens
- OPENAI_API_KEY: (optional) initial provider API key to seed a Channel

CI Behavior
- On push to main the workflow builds the Docker image and pushes to GCR (gcr.io/${{ secrets.GCP_PROJECT }}/one-api:${{ github.sha }}) then runs gcloud run deploy.

Steps to enable Cloud Run deploy
1. Create a Google Cloud project (if not present) and enable Cloud Run, Cloud Build, and Artifact Registry APIs.
2. Create a service account and grant roles: Cloud Run Admin, Service Account User, Storage Admin (or Artifact Registry Writer).
3. Create a JSON key for the service account and add it to your GitHub repository secrets as GCP_SA_KEY (use the raw JSON). Set GCP_PROJECT and CLOUD_RUN_REGION as secrets.
4. Add ONEAPI_JWT_SECRET and OPENAI_API_KEY as repository secrets (or inject provider keys via Secret Manager).
5. Push to main — workflow will run, build image, and deploy.

Local dev with Colima (mac)
1. Run scripts/colima-setup.sh to install colima & docker (requires Homebrew).
2. In repo root: cp .env.example .env and set ONEAPI_JWT_SECRET and optionally OPENAI_API_KEY. Example:
   ONEAPI_JWT_SECRET=$(head -c32 /dev/urandom | base64)
   OPENAI_API_KEY=sk-xxx
3. docker compose pull || true
4. docker compose up -d
5. Open http://localhost:3000, add Channels and Tokens via admin UI, then test API using the generated token.

1Password integration (recommended)
- Use 1Password CLI (op) to retrieve provider keys during local runs or inject into CI.
- Example (local):
  export OP_SERVICE_ACCOUNT_TOKEN=...
  op whoami
  # fetch item fields safely, then write to .env (do not echo values)

Notes
- This repo already includes Dockerfile and docker-compose.yml. The workflow expects to push to gcr.io; change tags if you prefer Artifact Registry or Docker Hub.
- For production, prefer storing provider keys in Secret Manager and referencing them from Cloud Run rather than embedding in env vars.
- After deploying, visit Google Cloud Console → Cloud Run to view service, logs, and revision details.
