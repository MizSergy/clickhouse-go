package tests

import (
	"context"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"127.0.0.1:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			//	Debug: true,
		})
	)
	if assert.NoError(t, err) {
		const ddl = `
			CREATE TABLE test_uuid (
				  Col1 UUID
				, Col2 UUID
				, Col3 Array(UUID)
				, Col4 Nullable(UUID)
				, Col5 Array(Nullable(UUID))
			) Engine Memory
		`
		if err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_uuid"); assert.NoError(t, err) {
			if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
				if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_uuid"); assert.NoError(t, err) {
					var (
						col1Data = uuid.New()
						col2Data = uuid.New()
					)
					if err := batch.Append(col1Data, col2Data, []uuid.UUID{col2Data, col1Data}, nil, []*uuid.UUID{
						&col1Data, nil, &col2Data,
					}); assert.NoError(t, err) {
						if assert.NoError(t, batch.Send()) {
							var (
								col1 uuid.UUID
								col2 uuid.UUID
								col3 []uuid.UUID
								col4 *uuid.UUID
								col5 []*uuid.UUID
							)
							if err := conn.QueryRow(ctx, "SELECT * FROM test_uuid").Scan(&col1, &col2, &col3, &col4, &col5); assert.NoError(t, err) {
								assert.Equal(t, col1Data, col1)
								assert.Equal(t, col2Data, col2)
								if assert.Nil(t, col4) {
									assert.Equal(t, []uuid.UUID{col2Data, col1Data}, col3)
									assert.Equal(t, []*uuid.UUID{
										&col1Data, nil, &col2Data,
									}, col5)
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestNullableUUID(t *testing.T) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"127.0.0.1:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			//Debug: true,
		})
	)
	if assert.NoError(t, err) {
		const ddl = `
			CREATE TABLE test_uuid (
				  Col1 Nullable(UUID)
				, Col2 Nullable(UUID)
			) Engine Memory
		`
		if err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_uuid"); assert.NoError(t, err) {
			if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
				if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_uuid"); assert.NoError(t, err) {
					var (
						col1Data = uuid.New()
						col2Data = uuid.New()
					)
					if err := batch.Append(col1Data, col2Data); assert.NoError(t, err) {
						if assert.NoError(t, batch.Send()) {
							var (
								col1 *uuid.UUID
								col2 *uuid.UUID
							)
							if err := conn.QueryRow(ctx, "SELECT * FROM test_uuid").Scan(&col1, &col2); assert.NoError(t, err) {
								assert.Equal(t, col1Data, *col1)
								assert.Equal(t, col2Data, *col2)
							}
						}
					}
				}
			}
		}
		if err := conn.Exec(ctx, "TRUNCATE TABLE test_uuid"); !assert.NoError(t, err) {
			return
		}
		if err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_uuid"); assert.NoError(t, err) {
			if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
				if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_uuid"); assert.NoError(t, err) {
					var col1Data = uuid.New()

					if err := batch.Append(col1Data, nil); assert.NoError(t, err) {
						if assert.NoError(t, batch.Send()) {
							var (
								col1 *uuid.UUID
								col2 *uuid.UUID
							)
							if err := conn.QueryRow(ctx, "SELECT * FROM test_uuid").Scan(&col1, &col2); assert.NoError(t, err) {
								if assert.Nil(t, col2) {
									assert.Equal(t, col1Data, *col1)
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestColumnarUUID(t *testing.T) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"127.0.0.1:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			//Debug: true,
		})
	)
	if assert.NoError(t, err) {
		const ddl = `
			CREATE TABLE test_uuid (
				  Col1 UUID
				, Col2 UUID
				, Col3 Nullable(UUID)
				, Col4 Array(UUID)
				, Col5 Array(Nullable(UUID))
			) Engine Memory
		`
		if err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_uuid"); assert.NoError(t, err) {
			if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
				if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_uuid"); assert.NoError(t, err) {
					var (
						col1Data []uuid.UUID
						col2Data []uuid.UUID
						col3Data []*uuid.UUID
						col4Data [][]uuid.UUID
						col5Data [][]*uuid.UUID
						v1, v2   = uuid.New(), uuid.New()
					)
					col1Data = append(col1Data, v1)
					col2Data = append(col2Data, v2)
					col3Data = append(col3Data, nil)
					col4Data = append(col4Data, []uuid.UUID{v1, v2})
					col5Data = append(col5Data, []*uuid.UUID{&v1, nil, &v2})
					for i := 0; i < 1000; i++ {
						if err := batch.Column(0).Append(col1Data); !assert.NoError(t, err) {
							return
						}
						if err := batch.Column(1).Append(col2Data); !assert.NoError(t, err) {
							return
						}
						if err := batch.Column(2).Append(col3Data); !assert.NoError(t, err) {
							return
						}
						if err := batch.Column(3).Append(col4Data); !assert.NoError(t, err) {
							return
						}
						if err := batch.Column(4).Append(col5Data); !assert.NoError(t, err) {
							return
						}
					}
					if assert.NoError(t, batch.Send()) {
						var (
							col1 uuid.UUID
							col2 uuid.UUID
							col3 *uuid.UUID
							col4 []uuid.UUID
							col5 []*uuid.UUID
						)
						if err := conn.QueryRow(ctx, "SELECT * FROM test_uuid LIMIT $1", 1).Scan(&col1, &col2, &col3, &col4, &col5); assert.NoError(t, err) {
							assert.Equal(t, v1, col1)
							assert.Equal(t, v2, col2)
							if assert.Nil(t, col3) {
								assert.Equal(t, []uuid.UUID{v1, v2}, col4)
								assert.Equal(t, []*uuid.UUID{&v1, nil, &v2}, col5)
							}
						}
					}
				}
			}
		}
	}
}
