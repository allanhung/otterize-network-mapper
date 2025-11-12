# Release Notes

## Version 25.11.1200 (2025-11-12)

### 🎉 New Features

#### MySQL Backend for External Traffic Storage

We're excited to announce persistent storage support for external traffic intents using MySQL! This major feature enables long-term tracking and analysis of your cluster's external traffic patterns.

**Key Capabilities:**
- **Persistent Storage**: Store external traffic intents in MySQL for historical analysis
- **GraphQL API**: Query external traffic data through a new `externalIntents` endpoint
- **Smart Filtering**: Automatically filters private IP addresses, storing only genuine external traffic
- **Performance Optimized**: Built-in local caching reduces database queries
- **GitHub Integration**: Optional webhook support for triggering CI/CD workflows when new external traffic is detected

### 📋 Configuration

Enable MySQL backend with environment variables:

```bash
# Database Configuration
OTTERIZE_DB_ENABLED=true              # Enable/disable MySQL storage
OTTERIZE_DB_HOST=mysql-host           # MySQL server host
OTTERIZE_DB_PORT=3306                 # MySQL server port
OTTERIZE_DB_USERNAME=root             # Database username
OTTERIZE_DB_PASSWORD=password         # Database password
OTTERIZE_DB_DATABASE=otterise         # Database name

# Optional: GitHub Actions Integration
OTTERIZE_GHA_DISPATCH_ENABLED=true    # Enable GitHub webhook
OTTERIZE_GHA_TOKEN=ghp_xxx            # GitHub personal access token
OTTERIZE_GHA_OWNER=your-org           # GitHub organization/user
OTTERIZE_GHA_REPO=your-repo           # Repository name
OTTERIZE_GHA_EVENT_TYPE=newIntent     # Event type for dispatch
```

### 🔌 GraphQL API

**New Query:**

```graphql
query GetExternalTraffic {
  externalIntents {
    client {
      name          # Service/Deployment name
      namespace     # Kubernetes namespace
      kind          # Resource type (Deployment, StatefulSet, etc.)
    }
    dnsName         # External DNS name accessed
    lastSeen        # Timestamp of last access (ISO 8601)
  }
}
```

### 🗄️ Database Schema

The MySQL backend automatically creates the following table:

```sql
CREATE TABLE external_traffic_intents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    client_name VARCHAR(128) NOT NULL,
    client_namespace VARCHAR(128) NOT NULL,
    client_kind VARCHAR(128) NOT NULL,
    dns_name VARCHAR(128) NOT NULL,
    last_seen DATE NOT NULL,
    UNIQUE KEY uniq_intent (client_name, client_namespace, client_kind, dns_name)
);
```

### 📦 Docker Images

New Dockerfiles are available for building custom images:
- `containers/Dockerfile.mapper` - Mapper component
- `containers/Dockerfile.sniffer` - Sniffer DaemonSet component

### 🔧 Technical Details

**Components Modified:**
- `mapper`: Added MySQL client initialization and GraphQL resolver
- `mysqlstore` package: New package for database operations
- GraphQL schema: Extended with `ExternalClient` and `ExternalIntent` types

**Dependencies Added:**
- `github.com/go-sql-driver/mysql v1.8.1` - MySQL driver

### ⚠️ Breaking Changes

None. The MySQL backend is optional and disabled by default. Existing deployments continue to work without any changes.

### 🐛 Bug Fixes

- Fixed schema generation issues in cloudclient package

### 📚 Documentation

- Updated README.md with MySQL configuration and usage examples
- Added GraphQL API documentation for external intents query

### 🙏 Acknowledgments

Thanks to all contributors who helped make this release possible!

---

## Upgrade Guide

### From v3.0.19 to v25.11.1200

**No action required** if you don't want to use the MySQL backend.

**To enable MySQL backend:**

1. Deploy a MySQL instance (v5.7+ or v8.0+) accessible from your Kubernetes cluster

2. Update your Helm values or deployment configuration:

   ```yaml
   mapper:
     env:
       - name: OTTERIZE_DB_ENABLED
         value: "true"
       - name: OTTERIZE_DB_HOST
         value: "mysql.default.svc.cluster.local"
       - name: OTTERIZE_DB_USERNAME
         valueFrom:
           secretKeyRef:
             name: mysql-credentials
             key: username
       - name: OTTERIZE_DB_PASSWORD
         valueFrom:
           secretKeyRef:
             name: mysql-credentials
             key: password
   ```

3. The mapper will automatically create the necessary tables on startup

4. Query external intents via GraphQL:
   ```bash
   curl -X POST http://mapper:9090/query \
     -H "Content-Type: application/json" \
     -d '{"query": "{ externalIntents { client { name namespace kind } dnsName lastSeen } }"}'
   ```

### Rollback

To disable MySQL backend, set `OTTERIZE_DB_ENABLED=false` or remove the environment variable entirely.

---

## Known Issues

- The Istio watcher still reports all HTTP traffic seen since sidecar startup (existing limitation)
- MySQL schema does not yet support storing pod-to-pod intents (future enhancement)

---

For questions, issues, or feature requests, please visit:
- GitHub Issues: https://github.com/otterize/network-mapper/issues
- Slack Community: https://joinslack.otterize.com
