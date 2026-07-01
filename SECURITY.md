# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |
| < 0.1   | :x:                |

The latest release is always recommended. Older 0.x releases may not receive backports unless the vulnerability is critical.

## Reporting a Vulnerability

We take the security of local-pvc-exporter seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please do NOT:

- Open a public GitHub issue
- Discuss the vulnerability in public forums
- Share the vulnerability with others until it has been resolved

### Please DO:

1. **Open a [private security advisory](https://github.com/alvarorg14/local-pvc-exporter/security/advisories/new)** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if you have one)

2. **Include the following information**:
   - Affected component(s)
   - Attack vector
   - Privileges required
   - User interaction required
   - CVSS score (if you can calculate it)

3. **Allow us 90 days** to address the vulnerability before public disclosure

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 7 days
- **Updates**: We will provide regular updates on the status of the vulnerability
- **Resolution**: We will work to resolve the issue as quickly as possible
- **Credit**: With your permission, we will credit you in our security advisories

### Security Best Practices

When deploying local-pvc-exporter:

1. **RBAC**: Use least-privilege RBAC policies; the chart grants read-only access to PVs, PVCs, nodes, and storage classes
2. **Network Policies**: Restrict network access to the metrics endpoint where possible
3. **Secrets**: Never commit secrets to version control
4. **Updates**: Keep local-pvc-exporter updated to the latest release
5. **Monitoring**: Monitor for suspicious activity and unexpected scrape errors
6. **Audit Logs**: Enable Kubernetes audit logging on clusters running the exporter

### Known Security Considerations

- **Host filesystem access**: The exporter mounts the host root read-only and walks PVC data directories to measure usage. It requires access to paths where volumes are stored on the node.
- **Container privileges**: Default Helm values run the container as root with only `CAP_DAC_READ_SEARCH` added to traverse directories with restrictive permissions (e.g. database data dirs). All other capabilities are dropped; `allowPrivilegeEscalation` is false and `readOnlyRootFilesystem` is true.
- **Kubernetes API access**: The exporter requires read-only access (`get`, `list`, `watch`) to persistent volumes, persistent volume claims, nodes, and storage classes.
- **Distroless image**: Production images are built with GoReleaser into a distroless base with no shell or package manager.
- **Metrics endpoint**: The `/metrics` HTTP endpoint exposes PVC usage data. Restrict access via network policies or ServiceMonitor configuration in production.

### Security Updates

Security updates will be:

- Released as patch versions (e.g., 0.2.2, 0.2.3)
- Announced via GitHub releases
- Tagged with `security` label where applicable

### Responsible Disclosure Timeline

- **Day 0**: Vulnerability reported
- **Day 1-2**: Acknowledgment and initial assessment
- **Day 3-7**: Detailed analysis and fix development
- **Day 8-30**: Testing and validation
- **Day 31-60**: Release preparation
- **Day 61-90**: Public disclosure (if not fixed earlier)

### Contact

For security-related issues, please open a [private security advisory](https://github.com/alvarorg14/local-pvc-exporter/security/advisories/new).

Thank you for helping keep local-pvc-exporter and its users safe!
