# Security Policy — Nexus Void

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 3.0.x   | :white_check_mark: |
| 2.0.x   | :x:                |
| < 2.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Nexus Void, please report it
responsibly. We take security seriously and will address issues promptly.

### How to Report

1. **Do NOT** open a public GitHub issue
2. Email **security@cybermindcli.com** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact assessment
   - Suggested fix (if any)

### Response Timeline

| Action | Timeline |
|--------|----------|
| Acknowledgment | Within 48 hours |
| Initial Assessment | Within 7 days |
| Fix Development | Within 30 days |
| Public Disclosure | After fix is released |

### Security Best Practices for Users

- Always run Nexus Void in isolated/test environments
- Keep the tool updated to the latest version
- Never expose the backend server to the public internet without authentication
- Use strong API keys and rotate them regularly
- Review all generated payloads before deployment
- Ensure compliance with local laws and authorization scope

### Built-in Security Features

- JWT-based authentication for dashboard access
- Rate limiting on API endpoints
- Input validation on all user-provided data
- SQLite with WAL mode for safe concurrent access
- CGO-disabled builds to minimize attack surface

---

*Created by Chandan Pandey | cybermindcli.com*
