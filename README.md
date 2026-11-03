# nombresmad

This repository contains a Go backend and React frontend for the NombresMad app.

## Google sign-in setup

Google proves the user's identity; the Go backend verifies the ID token and issues the app's own 12-hour bearer token only to Google accounts with a verified email listed in `AUTHORIZED_EMAILS`. No Google client secret is needed for this browser ID-token flow.

### Configure Google Cloud

1. Open the [Google Cloud Console](https://console.cloud.google.com/) and create a project or select the project for this app.
2. Open **Google Auth Platform** and configure **Branding**. Choose the audience:
   - Use **External** for personal Gmail accounts.
   - While the consent configuration is in **Testing**, add every intended Google account as a test user.
3. Create an OAuth client with application type **Web application**.
4. Add these **Authorized JavaScript origins** (origins only; no redirect URI is needed):
   - `http://localhost:3000`
   - Your exact production frontend origin, for example `https://your-app.up.railway.app`

   An origin contains only the scheme and host: do not include a path such as `/lista`, a trailing route, or a wildcard.
5. Copy the generated client ID ending in `.apps.googleusercontent.com`. Use this same client ID in the frontend and backend. Do not create or configure a client secret for this flow.
6. During development, keep the consent configuration in **Testing** and include the intended Google accounts as test users. Publish the consent configuration when ready for broader use, following Google's verification requirements if applicable. This sign-in requests basic identity information only; it does not grant access to Gmail or Drive data.

### Local Docker Compose

Create an untracked `.env` file in the repository root (it is ignored by Git) and replace every placeholder:

```dotenv
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
ADMIN_SECRET=replace-with-output-from-openssl-rand-base64-48
ADMIN_USER=admin@example.com
ADMIN_PASSWORD=replace-with-a-unique-high-entropy-password
AUTHORIZED_EMAILS=admin@example.com,another@example.com
```

Generate the signing secret with:

```bash
openssl rand -base64 48
```

Keep `ADMIN_SECRET` stable across restarts and deployments. Rotating it logs out every active app session. `ADMIN_PASSWORD` is the emergency password fallback; use a unique, high-entropy value even if you expect to sign in with Google. The fallback can be disabled by setting `ADMIN_PASSWORD` to an empty value. Never commit real secrets or passwords.

Start the local services:

```bash
docker-compose up --build
```

The frontend is available at `http://localhost:3000` and the API at `http://localhost:8080`. `GOOGLE_CLIENT_ID` is supplied to the backend as `GOOGLE_CLIENT_ID` and to the frontend build/runtime as `REACT_APP_GOOGLE_CLIENT_ID`. `AUTHORIZED_EMAILS` is a comma-separated list of exact email addresses permitted to sign in, including `ADMIN_USER` if password fallback is enabled. The backend fails to start if `GOOGLE_CLIENT_ID`, `ADMIN_SECRET`, or a valid non-empty `AUTHORIZED_EMAILS` list is missing. Emails are matched case-insensitively; a change to the list immediately invalidates sessions for removed addresses after restart.

For running services outside Docker, configure the same environment variables. From `backend`, run `go run ./cmd`; from `frontend`, run `npm install` and `npm start`. Run `./import.sh` from `backend` when the database needs its initial data.

Frontend/API configuration:

- `REACT_APP_API_URL` sets the browser-facing API base URL. It defaults to `/api`, keeping local requests on the frontend origin.
- `REACT_APP_GOOGLE_CLIENT_ID` is the Web OAuth client ID and must match backend `GOOGLE_CLIENT_ID`.
- `API_PROXY_TARGET` sets the frontend's API proxy target. It defaults to `http://localhost:8080`; Docker Compose sets it to `http://backend:8080`.

### Railway deployment

Configure these variables on the backend service:

```text
GOOGLE_CLIENT_ID=<same Web OAuth client ID used by frontend>
ADMIN_SECRET=<output from openssl rand -base64 48>
ADMIN_USER=<emergency admin email>
ADMIN_PASSWORD=<unique high-entropy emergency password>
AUTHORIZED_EMAILS=<comma-separated list of permitted email addresses>
```

Set `REACT_APP_GOOGLE_CLIENT_ID` to the same client ID on the frontend service. Also set `API_PROXY_TARGET` on the frontend service to `http://<backend-private-domain>:<backend-port>` using the backend's Railway private domain and assigned `PORT`; do not append `/api`. Add the exact production frontend origin to the Google OAuth client's authorized JavaScript origins.

Configure variables in Railway's service settings, then redeploy the backend and rebuild/redeploy the frontend so its build receives `REACT_APP_GOOGLE_CLIENT_ID`. Keep `ADMIN_SECRET` persistent and never commit real secrets or passwords. The backend listens on Railway's `PORT` and defaults to `8080` locally.

### Authentication verification checklist

- Only a Google account with a valid ID token, verified email, and an address in `AUTHORIZED_EMAILS` can sign in.
- The emergency password login works with the configured fallback credentials (or is unavailable when `ADMIN_PASSWORD` is empty).
- Logging out clears the local bearer token; expired sessions are rejected.
- Protected write operations reject requests without a currently valid bearer token.
