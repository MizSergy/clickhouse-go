package tests

import (
	"context"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAggregateFunction(t *testing.T) {
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
		if err := checkMinServerVersion(conn, 21, 1); err != nil {
			t.Skip(err.Error())
			return
		}
		const ddl = `
		CREATE TABLE test_simple_aggregate_function (
			  Col1 UInt64
			, Col2 SimpleAggregateFunction(sum, Double)
			, Col3 SimpleAggregateFunction(sumMap, Tuple(Array(Int16), Array(UInt64)))
		) Engine Memory
		`
		if err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_simple_aggregate_function"); assert.NoError(t, err) {
			if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
				if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_simple_aggregate_function"); assert.NoError(t, err) {
					var (
						col1Data = uint64(42)
						col2Data = float64(256.1)
						col3Data = []interface{}{
							[]int16{1, 2, 3, 4, 5},
							[]uint64{1, 2, 3, 4, 5},
						}
					)
					if err := batch.Append(col1Data, col2Data, col3Data); !assert.NoError(t, err) {
						return
					}
					if assert.NoError(t, batch.Send()) {
						var result struct {
							Col1 uint64
							Col2 float64
							Col3 []interface{}
						}
						if err := conn.QueryRow(ctx, "SELECT * FROM test_simple_aggregate_function").ScanStruct(&result); assert.NoError(t, err) {
							assert.Equal(t, col1Data, result.Col1)
							assert.Equal(t, col2Data, result.Col2)
							assert.Equal(t, col3Data, result.Col3)
						}
					}
				}
			}
		}
	}
}
