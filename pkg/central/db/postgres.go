package db

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"strconv"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq" // add postgresql driver
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/shared"
)

// DB .
type DB struct {
	*sql.DB
	cache squirrel.DBProxyBeginner // statement cache

}

type DBClient interface {

	// internal house keeping
	GetLastProcessedSampleID() (uint64, error)
	SetLastProcessedSampleID(id uint64) error

	// client/server api
	InsertSample(s Sample) error
	InsertSimpleSample(s SimpleSample) error
	IsURLAllowed(url *url.URL, countryCode string) (bool, error)
	RecentSuggestionSessions(n uint64) ([]tokenData, error)
	GetSamples(fromID uint64, sampleType string) (chan Sample, error)
	PublishHost(sample Sample) error
	GetBlockedHosts(CountryCode string, ASN int) ([]string, error)
	GetRelatedHosts() (map[string][]string, error)
	GetUpgrade(GetUpgradeQuery) (UpgradeMeta, bool, error)
	InsertUpgrades([]UpgradeMeta) error

	// statstics export api

	// Returns credentials for given user, if exists
	GetExportAPIAuthCredentials(username string) (bool, APICredentials, error)
	// create or update credentials, does not enable if disabled.
	InsertExportAPICredentials(credentials APICredentials) error

	// query for exporting data from logs...
	GetExportBlockedHosts(req shared.BlockedContentRequest) ([]shared.HostsPublishLog, string, error)
	GetExportSamples(req shared.ExportSampleRequest) ([]shared.ExportSampleEntry, string, error)
	GetExportSimpleSamples(req shared.ExportSimpleSampleRequest) ([]shared.ExportSimpleSampleEntry, string, error)
}

// Sample mirrors the samples postgres table
type Sample struct {
	ID          uint64
	Host        string
	CountryCode string
	ASN         int
	CreatedAt   time.Time
	Origin      string
	Type        string
	Token       shared.SuggestionToken
	Data        []byte
	ExtraData   []byte
}

// Sample mirrors the samples postgres table
type SimpleSample struct {
	ID          uint64
	CountryCode string
	ASN         int
	CreatedAt   time.Time
	Type        string
	OriginID    string
	Data        []byte
}

// HostListEntry .
type HostListEntry struct {
	ID          uint64
	Host        string
	CountryCode string
	ASN         int
	CreatedAt   time.Time
	Sticky      bool
}

// UpgradeMeta .
type UpgradeMeta struct {
	Artifact         string    `json:"artifact"`
	Version          string    `json:"version"`
	CreatedAt        time.Time `json:"createdAt"`
	SHA256Sum        string    `json:"sha256Sum"`
	ED25519Signature string    `json:"ed25519Sig"`
}

// Open returns a wrapped *sql.DB and starts services
func Open(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return &DB{db, squirrel.NewStmtCacheProxy(db)}, nil
}

func logSQLErr(err error, query squirrel.Sqlizer) {
	if err != nil {
		if lg.V(10) {
			sql, args, qerr := query.ToSql()
			var msg string
			if qerr != nil {
				msg = fmt.Sprintf("sql error: %s", err)
			} else {
				msg = fmt.Sprintf("sql error: %s: %+v, %s", sql, args, err)
			}
			lg.ErrorDepth(1, msg)
		} else {
			lg.ErrorDepth(1, "sql error: %s", err.Error())
		}
	} else if lg.V(19) {
		sql, _, qerr := query.ToSql()
		if qerr != nil {
			lg.ErrorDepth(1, fmt.Sprintf("sql error: %s", err))

		}
		lg.InfoDepth(1, fmt.Sprintf("sql query: %s", sql))
	}
}

const PageLength = 1000

func (d *DB) GetExportBlockedHosts(req shared.BlockedContentRequest) ([]shared.HostsPublishLog, string, error) {
	var results []shared.HostsPublishLog

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	i := psql.
		Select("id", "host", "country_code", "asn", "created_at", "sticky", "action").
		From("hosts_publish_log").
		OrderBy("id desc").
		Limit(PageLength + 1)
	if req.IDMax != 0 {
		i = i.Where("id < ?", req.IDMax)
	}
	rows, err := i.RunWith(d.cache).Query()
	if err != nil {
		logSQLErr(err, &i)
		return nil, "", err
	}
	defer rows.Close()
	next := ""
	count := 0
	for rows.Next() {
		var item shared.HostsPublishLog
		err := rows.Scan(
			&item.ID,
			&item.Host,
			&item.CountryCode,
			&item.ASN,
			&item.CreatedAt,
			&item.Sticky,
			&item.Action,
		)
		count++
		if err != nil {
			lg.Warning(err)
			continue
		}
		if count > PageLength {
			next = item.ID
		} else {
			results = append(results, item)
		}
	}
	return results, next, nil
}

func (d *DB) GetExportSamples(req shared.ExportSampleRequest) ([]shared.ExportSampleEntry, string, error) {
	var results []shared.ExportSampleEntry

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	i := psql.
		Select("id", "host", "country_code", "asn", "created_at", "origin", "type", "token", "data", "extra_data").
		From("samples").
		OrderBy("id desc").
		Limit(PageLength + 1)
	if req.IDMax != 0 {
		i = i.Where("id < ?", req.IDMax)
	}
	rows, err := i.RunWith(d.cache).Query()
	if err != nil {
		logSQLErr(err, &i)
		return nil, "", err
	}
	defer rows.Close()
	next := ""
	count := 0
	for rows.Next() {
		var i shared.ExportSampleEntry
		err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.CountryCode,
			&i.ASN,
			&i.CreatedAt,
			&i.Origin,
			&i.Type,
			&i.Token,
			&i.Data,
			&i.ExtraData,
		)
		count++
		if err != nil {
			lg.Warning(err)
			continue
		}
		if count > PageLength {
			next = i.ID
		} else {
			results = append(results, i)
		}
	}
	return results, next, nil
}

func (d *DB) GetExportSimpleSamples(req shared.ExportSimpleSampleRequest) ([]shared.ExportSimpleSampleEntry, string, error) {
	var results []shared.ExportSimpleSampleEntry

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	i := psql.
		Select("id", "country_code", "asn", "created_at", "type", "origin_id", "data").
		From("simple_samples").
		OrderBy("id desc").Limit(PageLength + 1)
	if req.IDMax != 0 {
		i = i.Where("id < ?", req.IDMax)
	}
	rows, err := i.RunWith(d.cache).Query()
	if err != nil {
		logSQLErr(err, &i)
		return nil, "", err
	}
	defer rows.Close()
	next := ""
	count := 0
	for rows.Next() {
		var i shared.ExportSimpleSampleEntry
		err := rows.Scan(
			&i.ID,
			&i.CountryCode,
			&i.ASN,
			&i.CreatedAt,
			&i.Type,
			&i.OriginID,
			&i.Data,
		)
		count++
		if err != nil {
			lg.Warning(err)
			continue
		}
		if count > PageLength {
			next = i.ID
		} else {
			results = append(results, i)
		}
	}
	return results, next, nil
}

func (d *DB) GetExportAPIAuthCredentials(username string) (bool, APICredentials, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	i := psql.
		Select("enabled", "username", "hash", "salt").
		From("export_api_auth").
		Where(squirrel.Eq{"username": username}).
		Limit(1)
	row := i.RunWith(d.cache).QueryRow()
	cred := APICredentials{}
	var hashstr, saltstr string
	err := row.Scan(&cred.Enabled, &cred.Username, &hashstr, &saltstr)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, APICredentials{}, nil
		default:
			return false, APICredentials{}, err
		}
	}

	hashdata, err := base64.StdEncoding.DecodeString(hashstr)
	if err != nil {
		return false, APICredentials{}, err
	}
	cred.PasswordHash = hashdata

	saltdata, err := base64.StdEncoding.DecodeString(saltstr)
	if err != nil {
		return false, APICredentials{}, err
	}
	cred.Salt = saltdata

	return true, cred, nil
}

func (d *DB) InsertExportAPICredentials(cred APICredentials) error {
	if cred.Salt == nil {
		return errors.New("Salt not set")
	}
	if cred.PasswordHash == nil {
		return errors.New("Passwordhash not set")
	}
	if cred.Username == "" {
		return errors.New("Username not set")
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	hashstr := base64.StdEncoding.EncodeToString(cred.PasswordHash)
	saltstr := base64.StdEncoding.EncodeToString(cred.Salt)

	i := psql.Insert("export_api_auth").
		Columns("username", "hash", "salt").
		Values(cred.Username, hashstr, saltstr)

	_, err := i.RunWith(d.cache).Exec()
	logSQLErr(err, &i)
	return err
}

// InsertSample inserts a Sample into the samples table.
func (d *DB) InsertSample(s Sample) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	columns := []string{
		"host", "country_code", "asn",
		"type", "origin",
		"token", "data"}
	var values []interface{}
	values = append(values,
		s.Host, s.CountryCode, s.ASN,
		s.Type, s.Origin, string(s.Token),
		s.Data)
	if s.ExtraData != nil {
		columns = append(columns, "extra_data")
		values = append(values, s.ExtraData)
	}

	i := psql.Insert("samples").Columns(columns...).Values(values...)
	_, err := i.RunWith(d.cache).Exec()
	logSQLErr(err, &i)
	return err
}

// InsertSample inserts a Sample into the samples table.
func (d *DB) InsertSimpleSample(s SimpleSample) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	columns := []string{"country_code", "asn", "type", "origin_id"}
	var values []interface{}
	values = append(values, s.CountryCode, s.ASN, s.Type, s.OriginID)
	if s.Data != nil {
		columns = append(columns, "data")
		values = append(values, s.Data)
	}
	i := psql.Insert("simple_samples").Columns(columns...).Values(values...)
	_, err := i.RunWith(d.cache).Exec()
	logSQLErr(err, &i)
	return err
}

func (d *DB) GetSamples(fromID uint64, sampleType string) (chan Sample, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	i := psql.
		Select("id", "host", "country_code", "asn", "created_at", "origin", "type", "data", "extra_data").
		From("samples").
		Where("id > ?", fromID)
	rows, err := i.RunWith(d.cache).Query()
	if err != nil {
		return nil, err
	}

	results := make(chan Sample, 0)
	go func(rows *sql.Rows) {
		defer rows.Close()
		defer close(results)
		for rows.Next() {
			var sample Sample
			err := rows.Scan(
				&sample.ID,
				&sample.Host,
				&sample.CountryCode,
				&sample.ASN,
				&sample.CreatedAt,
				&sample.Origin,
				&sample.Type,
				&sample.Data,
				&sample.ExtraData,
			)
			if err != nil {
				lg.Warning(err)
				continue
			}
			results <- sample
		}
	}(rows)
	return results, nil
}

func (d *DB) GetBlockedHosts(CountryCode string, ASN int) ([]string, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	s := psql.Select("host").From("hosts_publish").Where(squirrel.Eq{
		"country_code": CountryCode,
		"asn":          ASN,
	})
	rows, err := s.RunWith(d.cache).Query()
	if err != nil {
		logSQLErr(err, &s)
		return nil, err
	}
	defer rows.Close()
	var hosts []string
	for rows.Next() {
		var host string
		err := rows.Scan(&host)
		if err != nil {
			lg.Fatal(err)
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (d *DB) GetRelatedHosts() (map[string][]string, error) {
	result := make(map[string][]string, 0)
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	s := psql.Select("host", "related").From("hosts_related")
	rows, err := s.RunWith(d.cache).Query()
	if err != nil {
		logSQLErr(err, &s)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var host, related string
		err := rows.Scan(&host, &related)
		if err != nil {
			lg.Fatal(err)
		}
		if relateds, ok := result[host]; ok {
			relateds = append(relateds, related)
			result[host] = relateds
		} else {
			relateds := make([]string, 1)
			relateds = append(relateds, related)
			result[host] = relateds
		}
	}
	return result, nil
}

// InsertUpgrades inserts a list of upgrade entreis into the database
func (d *DB) InsertUpgrades(u []UpgradeMeta) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	tx, err := d.cache.Begin()
	if err != nil {
		return err
	}

	for _, v := range u {
		i := psql.Insert("upgrades").
			Columns("artifact", "version", "sha256sum", "ed25519sig").
			Values(v.Artifact, v.Version, v.SHA256Sum, v.ED25519Signature)
		_, err := i.RunWith(d.cache).Exec()
		if err != nil {
			logSQLErr(err, &i)
			if err := tx.Rollback(); err != nil {
				lg.Errorln(err)
			}
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		lg.Errorln(err)
	}
	return nil
}

// GetUpgradeQuery .
type GetUpgradeQuery struct {
	Artifact        string
	Version         string
	AlsoUnpublished bool // include unpublished versions in results
}

func (d *DB) GetUpgrade(q GetUpgradeQuery) (UpgradeMeta, bool, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	wh := squirrel.Eq{"artifact": q.Artifact}

	if !q.AlsoUnpublished {
		wh["published"] = true
	}

	if q.Version != "" {
		wh["version"] = q.Version
	}

	var result UpgradeMeta
	s := psql.
		Select("artifact", "version", "created_at", "sha256sum", "ed25519sig").
		From("upgrades").
		Where(wh).
		Limit(1)

	row := s.RunWith(d.cache).QueryRow()
	err := row.Scan(
		&result.Artifact,
		&result.Version,
		&result.CreatedAt,
		&result.SHA256Sum,
		&result.ED25519Signature,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, false, nil
		default:
			return result, false, err
		}
	}
	return result, true, nil
}

func (d *DB) PublishHost(sample Sample) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	s := psql.Select("1").From("hosts_publish").Where(squirrel.Eq{
		"host":         sample.Host,
		"country_code": sample.CountryCode,
		"asn":          sample.ASN,
	}).Limit(1).Prefix("select exists(").Suffix(")")

	row := s.RunWith(d.cache).QueryRow()

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		logSQLErr(err, &s)
		return err
	}
	if !exists {
		i := psql.Insert("hosts_publish").
			Columns("host", "country_code", "asn").
			Values(sample.Host, sample.CountryCode, sample.ASN)
		_, err := i.RunWith(d.cache).Exec()
		if err != nil {
			logSQLErr(err, &i)
			return err
		}
	}
	return nil
}

// IsURLAllowed returns true if the supplied URL is supported for circumenvtion with alkasir.
func (d *DB) IsURLAllowed(url *url.URL, countryCode string) (bool, error) {
	// TODO: Also needs to match ASN
	var isUnsupported bool
	err := d.QueryRow(
		"select exists(select 1 from hosts_unsupported where host = $1 and country_code = $2)",
		url.Host, countryCode,
	).Scan(&isUnsupported)
	return !isUnsupported, err
}

func (d *DB) RecentSuggestionSessions(n uint64) ([]tokenData, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	q := psql.
		Select("token", "created_at", "data::json->'URL'").
		From("samples").
		Where(squirrel.Eq{"type": "NewClientToken"}).
		OrderBy("ID desc").
		Limit(n)

	rows, err := q.RunWith(d.cache).Query()
	if err != nil {
		lg.Fatal(err)
	}
	defer rows.Close()
	var tokens []tokenData
	for rows.Next() {
		var token tokenData
		var ID string
		err := rows.Scan(&ID, &token.CreatedAt, &token.URL)
		if err != nil {
			lg.Fatal(err)
		}
		token.ID = shared.SuggestionToken(ID)
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// kv .
type kv struct {
	*DB
	Table string
	Key   string
}

func (d *DB) GetLastProcessedSampleID() (uint64, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	s := psql.Select("value").From("central_state").Where(
		squirrel.Eq{"name": "last_processed_sample_id"},
	).Limit(1)

	row := s.RunWith(d.cache).QueryRow()
	var valuestr string
	err := row.Scan(&valuestr)
	if err != nil {
		if err == sql.ErrNoRows {
			lg.Warning("last processed sample id not set, returning 0 (default value)")
			return 0, nil
		}
		return 0, err
	}
	v, err := strconv.ParseUint(valuestr, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (d *DB) SetLastProcessedSampleID(id uint64) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	idstr := strconv.FormatUint(id, 10)

	s := psql.
		Update("central_state").
		Set("value", idstr).
		Where(squirrel.Eq{"name": "last_processed_sample_id"})
	sr, err := s.RunWith(d.cache).Exec()
	if err != nil {
		logSQLErr(err, &s)
		log.Fatal(err)
	}
	sn, err := sr.RowsAffected()
	if err != nil {
		lg.Warningln(err)
		return err
	}
	if sn > 0 {
		return nil
	}

	i := psql.
		Insert("central_state").
		Columns("name", "value").
		Values("last_processed_sample_id", idstr)

	ir, err := i.RunWith(d.cache).Exec()
	if err != nil {
		logSQLErr(err, &i)
		log.Fatal(err)
	}
	in, err := ir.RowsAffected()
	if err != nil {
		lg.Warningln(err)
		return err
	}
	if in > 0 {
		return nil
	}
	log.Fatal("unknown error")
	return nil
}
