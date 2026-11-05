# Contract: simulator admin API (both flavours)

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| GET | `/_sim/healthz` | — | 200 `{"status":"ok","flavor":…}` |
| GET | `/_sim/faults` | — | 200 active FaultConfig (outage with absolute `start`/`end`) |
| PUT | `/_sim/faults` | FaultConfig JSON | 200 new config; 400 on validation error |
| PATCH | `/_sim/faults` | partial FaultConfig | 200 merged config |
| DELETE | `/_sim/faults` | — | 200 start-up config restored |
| GET | `/_sim/requests` | — | 200 `{"requests":[RequestRecord…]}` |
| DELETE | `/_sim/requests` | — | 204 |

Admin endpoints are never subject to faults and are not recorded in the request log.
