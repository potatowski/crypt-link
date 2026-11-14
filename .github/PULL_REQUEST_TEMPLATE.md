# 🚀 Pull Request Template

## 📝 Summary

*Briefly describe the "what" and the "why." What logic is being introduced or modified? Why was this specific approach taken over alternatives?*

## 🔗 Context & Impact

* **Related Issue:** #
* **API Changes:** Does this change the API contract? (Yes/No)
* **Breaking Change:** Does this require coordinated frontend/mobile updates? (Yes/No)
* **Dependencies:** Were any new packages added? (List them if yes)

## 🛠️ Implementation Details

*Detail any specific architectural decisions, new dependencies, or "gotchas" the reviewer should know about.*
> **Note:** Mention if any database migrations or environment variable changes are required.

### Database & Infrastructure

* [ ] **Migrations:** Included and tested for both `up` and `down` (rollback).
* [ ] **Indexes:** New queries are optimized with appropriate indexing.
* [ ] **Env Vars:** New keys needed in `.env` (List them: `...`)

### Logic & Security

* [ ] **Error Handling:** Edge cases (404s, 500s, timeouts) are handled.
* [ ] **Auth/Authz:** Proper permissions/scopes are checked for this endpoint.
* [ ] **Validation:** Input data is sanitized and validated.

---

## 🧪 Testing & Validation

### Automated Tests

* [ ] **Unit Tests:** Added for core business logic.
* [ ] **Integration Tests:** API endpoints verified with a test database.
* **Coverage:** Current coverage percentage or delta: `XX%`

### Manual Verification (API Request/Response)

*Provide a sample CURL command or a snippet of the JSON response for the reviewer to verify.*

```bash
# Example CURL for testing
curl -X GET "http://localhost:8080/api/v1/resource" -H "Authorization: Bearer <token>"

```

---

## 📈 Performance & Monitoring

* [ ] **Logging:** Significant events or failures are logged.
* [ ] **Scaling:** How does this perform with  records? (e.g., avoided  query issues).
* [ ] **Alerting:** (If applicable) New metrics or alerts have been configured.
---

## 📋 Changelog Entry

### Added
- 
### Changed
- 
### Deprecated
-
### Removed
-
### Fixed
-
### Security
-
