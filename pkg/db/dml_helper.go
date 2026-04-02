package db

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func BuildRepeatedObjectName(objectName string, index string) string {
	return fmt.Sprintf("%s_%s", objectName, index)
}

func BuildPKConstraintName(tableName string) string {
	return fmt.Sprintf("%s_pk", tableName)
}

func BuildUKConstraintName(tableName string, fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return fmt.Sprintf("%s_uk", tableName)
	}

	return fmt.Sprintf("%s_%s_uk", tableName, buildFieldNamesHash(fieldNames...))
}

func BuildFKConstraintName(tableName string, fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return fmt.Sprintf("%s_fk", tableName)
	}

	return fmt.Sprintf("%s_%s_fk", tableName, buildFieldNamesHash(fieldNames...))
}

func buildFieldNamesHash(fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return ""
	}
	if len(fieldNames) == 1 {
		return fieldNames[0]
	}

	builder := strings.Builder{}
	for _, field := range fieldNames {
		builder.WriteString(field)
	}
	hasher := md5.New()
	hasher.Write([]byte(builder.String()))

	return hex.EncodeToString(hasher.Sum(nil))
}
