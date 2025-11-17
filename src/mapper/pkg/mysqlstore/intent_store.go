package mysqlstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/otterize/network-mapper/src/mapper/pkg/externaltrafficholder"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

type MySQLIntentStore struct {
	Db                  *sql.DB
	config              Config
	localIntentCacheMap map[string]map[string]struct{}
}

type ExternalTrafficPayload struct {
	ClientName      string `json:"client_name"`
	ClientNamespace string `json:"client_namespace"`
	ClientKind      string `json:"client_kind"`
	DNSName         string `json:"dns_name"`
}

func NewMySQLIntentStore(config Config) (*MySQLIntentStore, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.DbUsername, config.DbPassword, config.DbHost, config.DbPort, config.DbDatabase)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.WithError(err).Error("failed to open database connection")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logrus.WithError(err).Error("failed to ping database")
		return nil, err
	}

	store := &MySQLIntentStore{
		Db:                  db,
		config:              config,
		localIntentCacheMap: make(map[string]map[string]struct{}),
	}

	err = store.ensureTableExists()
	if err != nil {
		logrus.WithError(err).Error("failed to ensure table exists")
		return nil, err
	}
	if err := store.LoadCacheFromDb(); err != nil {
		logrus.WithError(err).Error("failed to load cache from db")
	}
	return store, nil
}

func (s *MySQLIntentStore) ensureTableExists() error {
	query := `
            CREATE TABLE IF NOT EXISTS external_traffic_intents (
                id BIGINT AUTO_INCREMENT PRIMARY KEY,
                client_name VARCHAR(128) NOT NULL,
                client_namespace VARCHAR(128) NOT NULL,
                client_kind VARCHAR(128) NOT NULL,
                dns_name VARCHAR(128) NOT NULL,
                last_seen DATE NOT NULL,
                UNIQUE KEY uniq_intent (client_name, client_namespace, client_kind, dns_name)
            )
        `

	_, err := s.Db.Exec(query)
	return err
}

func (s *MySQLIntentStore) LoadCacheFromDb() error {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	rows, err := s.Db.Query(`
            SELECT client_name, client_namespace, client_kind, dns_name, last_seen
              FROM external_traffic_intents
             WHERE last_seen >= ?
               AND last_seen <= ?
        `, today, yesterday)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var clientName, clientNamespace, clientKind, dnsName string
		var lastSeen time.Time

		if err := rows.Scan(&clientName, &clientNamespace, &clientKind, &dnsName, &lastSeen); err != nil {
			return err
		}

		date := lastSeen.Format("2006-01-02")
		if _, ok := s.localIntentCacheMap[date]; !ok {
			s.localIntentCacheMap[date] = make(map[string]struct{})
		}

		cacheKey := fmt.Sprintf("%s|%s|%s|%s", clientName, clientNamespace, clientKind, dnsName)
		s.localIntentCacheMap[date][cacheKey] = struct{}{}
	}

	return rows.Err()
}

func (s *MySQLIntentStore) LogExternalTrafficIntentsCallback(ctx context.Context, intents []externaltrafficholder.TimestampedExternalTrafficIntent) {

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	dayBeforeYesterday := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	delete(s.localIntentCacheMap, dayBeforeYesterday)

	if _, ok := s.localIntentCacheMap[today]; !ok {
		s.localIntentCacheMap[today] = make(map[string]struct{})
	}

	hasNewIntent := false
	var payloads []ExternalTrafficPayload
	ignoreClients := strings.Split(s.config.ClientIgnoreListByName, ",")
	ignoreClientSet := make(map[string]struct{}, len(ignoreClients))
	for _, item := range ignoreClients {
		ignoreClientSet[strings.TrimSpace(item)] = struct{}{}
	}

	for _, ti := range intents {
		if hasExternalIP(ti.Intent.IPs) {
			if _, ignored := ignoreClientSet[ti.Intent.Client.Name]; !ignored {
				exists := s.storeIntent(ctx, ti, today, yesterday)
				if !exists {
					hasNewIntent = true
					printLog(ti, "info", "Received new intent")
					clientKind := ""
					if ti.Intent.Client.PodOwnerKind != nil {
						clientKind = ti.Intent.Client.PodOwnerKind.Kind
					}

					payload := ExternalTrafficPayload{
						ClientName:      ti.Intent.Client.Name,
						ClientNamespace: ti.Intent.Client.Namespace,
						ClientKind:      clientKind,
						DNSName:         ti.Intent.DNSName,
					}
					payloads = append(payloads, payload)
				}
				printLog(ti, "debug", "Received external traffic intent")
			}
		}
	}
	if hasNewIntent && s.config.GhaDispatchEnabled {
		s.dispatchGithubAction(ctx, payloads)
	}
}

func hasExternalIP(ips map[externaltrafficholder.IP]struct{}) bool {
	privateCIDRs := []string{
		"127.0.0.1/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
	}

	var privateNets []*net.IPNet
	for _, cidr := range privateCIDRs {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			logrus.WithError(err).Errorf("Warning: invalid CIDR %s, skipping", cidr)
			continue
		}
		privateNets = append(privateNets, block)
	}

	for ipRaw := range ips {
		ipStr := string(ipRaw)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}

		isPrivate := false
		for _, privateNet := range privateNets {
			if privateNet.Contains(ip) {
				isPrivate = true
				break
			}
		}

		if !isPrivate {
			return true // Found an external IP
		}
	}

	return false
}

func printLog(ti externaltrafficholder.TimestampedExternalTrafficIntent, logLevel, logMessage string) {
	clientKind := ""
	if ti.Intent.Client.PodOwnerKind != nil && ti.Intent.Client.PodOwnerKind.Kind != "" {
		clientKind = ti.Intent.Client.PodOwnerKind.Kind
	}
	entry := logrus.WithFields(logrus.Fields{
		"timestamp":        ti.Timestamp.Format("2006-01-02"),
		"client_name":      ti.Intent.Client.Name,
		"client_namespace": ti.Intent.Client.Namespace,
		"client_kind":      clientKind,
		"DnsName":          ti.Intent.DNSName,
	})

	switch strings.ToLower(logLevel) {
	case "trace":
		entry.Trace(logMessage)
	case "info":
		entry.Info(logMessage)
	case "warn":
		entry.Warn(logMessage)
	case "error":
		entry.Error(logMessage)
	case "fatal":
		entry.Fatal(logMessage)
	case "panic":
		entry.Panic(logMessage)
	default:
		entry.Debug(logMessage)
	}
}

func (s *MySQLIntentStore) dispatchGithubAction(ctx context.Context, payloads []ExternalTrafficPayload) {

	url := fmt.Sprintf("https://%s/repos/%s/%s/dispatches", s.config.GhaUrl, s.config.GhaOwner, s.config.GhaRepo)

	payload := map[string]interface{}{
		"event_type": fmt.Sprintf("%s-%s", s.config.Cluster, s.config.GhaEventType),
		"client_payload": map[string]interface{}{
			"cluster": s.config.Cluster,
			"intents": payloads,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal payload")
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithError(err).Error("failed to do http post")
		return
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.GhaToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("failed to get http response")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		logrus.Info("Dispatch event triggered successfully")
	} else {
		logrus.WithFields(logrus.Fields{"status": resp.Status}).Error("Failed to trigger dispatch")
	}
}

func (s *MySQLIntentStore) checkIfExists(ctx context.Context, clientName, clientNamespace, clientKind, dnsName, yesterday string) (found bool) {
	cacheKey := fmt.Sprintf("%s|%s|%s", clientName, clientNamespace, dnsName)

	if _, existedYesterday := s.localIntentCacheMap[yesterday][cacheKey]; existedYesterday {
		return true
	}
	var exists bool
	err := s.Db.QueryRowContext(ctx, `
		    SELECT EXISTS(
		      SELECT 1 FROM external_traffic_intents
		       WHERE client_name = ? AND client_namespace = ? AND client_kind = ? AND dns_name = ?
		    )
	      `, clientName, clientNamespace, clientKind, dnsName).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("failed to query db for cache")
		return false
	}

	return exists
}

func (s *MySQLIntentStore) storeIntent(ctx context.Context, ti externaltrafficholder.TimestampedExternalTrafficIntent, today, yesterday string) (found bool) {
	insertSql := `
            INSERT INTO external_traffic_intents (client_name, client_namespace, client_kind, dns_name, last_seen)
            VALUES (?, ?, ?, ?, ?)
        `
	updateSql := `
            UPDATE external_traffic_intents 
               SET last_seen = ?
             WHERE client_name = ? AND client_namespace = ? AND client_kind = ? AND dns_name = ?
        `

	intent := ti.Intent
	intentDate := ti.Timestamp.Format("2006-01-02")

	clientKind := ""
	if intent.Client.PodOwnerKind != nil && intent.Client.PodOwnerKind.Kind != "" {
		clientKind = intent.Client.PodOwnerKind.Kind
	}
	cacheKey := fmt.Sprintf("%s|%s|%s|%s", intent.Client.Name, intent.Client.Namespace, clientKind, intent.DNSName)

	if _, exists := s.localIntentCacheMap[today][cacheKey]; exists {
		logrus.Debug("cache hit: today, skipping insert")
		return exists
	}

	intentExists := s.checkIfExists(ctx, intent.Client.Name, intent.Client.Namespace, clientKind, intent.DNSName, yesterday)
	if intentExists {
		_, err := s.Db.ExecContext(ctx, updateSql,
			intentDate,
			intent.Client.Name,
			intent.Client.Namespace,
			clientKind,
			intent.DNSName,
		)
		if err != nil {
			logrus.WithError(err).Error("failed to update intent")
			return intentExists
		}
		s.localIntentCacheMap[today][cacheKey] = struct{}{}
		return intentExists
	}

	_, err := s.Db.ExecContext(ctx, insertSql,
		intent.Client.Name,
		intent.Client.Namespace,
		clientKind,
		intent.DNSName,
		intentDate,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to insert intent")
		return intentExists
	}
	s.localIntentCacheMap[today][cacheKey] = struct{}{}
	return intentExists
}

// ExternalIntentRecord represents a record from the database
type ExternalIntentRecord struct {
	ClientName      string
	ClientNamespace string
	ClientKind      string
	DNSName         string
	LastSeen        time.Time
}

// CleanupExpiredIntents removes intents older than the configured retention period
func (s *MySQLIntentStore) CleanupExpiredIntents(ctx context.Context) error {
	if s.config.RetentionDays <= 0 {
		logrus.Debug("Retention cleanup skipped: retention days not configured or invalid")
		return nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)
	query := `
		DELETE FROM external_traffic_intents
		WHERE last_seen < ?
	`

	result, err := s.Db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		logrus.WithError(err).Error("failed to cleanup expired intents")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Warn("failed to get rows affected count")
	} else if rowsAffected > 0 {
		logrus.WithFields(logrus.Fields{
			"rows_deleted":   rowsAffected,
			"retention_days": s.config.RetentionDays,
			"cutoff_date":    cutoffDate.Format("2006-01-02"),
		}).Info("Cleaned up expired external traffic intents")
	}

	return nil
}

// GetExternalIntents retrieves all external traffic intents from the database
func (s *MySQLIntentStore) GetExternalIntents(ctx context.Context) ([]ExternalIntentRecord, error) {
	// Cleanup expired intents before querying
	if err := s.CleanupExpiredIntents(ctx); err != nil {
		logrus.WithError(err).Warn("failed to cleanup expired intents, continuing with query")
	}

	query := `
		SELECT client_name, client_namespace, client_kind, dns_name, last_seen
		FROM external_traffic_intents
		ORDER BY last_seen DESC
	`

	rows, err := s.Db.QueryContext(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("failed to query external intents")
		return nil, err
	}
	defer rows.Close()

	var intents []ExternalIntentRecord
	for rows.Next() {
		var record ExternalIntentRecord
		if err := rows.Scan(
			&record.ClientName,
			&record.ClientNamespace,
			&record.ClientKind,
			&record.DNSName,
			&record.LastSeen,
		); err != nil {
			logrus.WithError(err).Error("failed to scan external intent row")
			continue
		}
		intents = append(intents, record)
	}

	if err := rows.Err(); err != nil {
		logrus.WithError(err).Error("error iterating external intent rows")
		return nil, err
	}

	return intents, nil
}
