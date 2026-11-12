package postgres

import (
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	extpg "github.com/ElfAstAhe/go-service-template/pkg/ext/repository/postgres"
	repo "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
)

// pgStatementsBuilder инициализирует и связывает SQL-запросы PostgreSQL с внутренними идентификаторами стейтментов.
func (dao *DataAuditOutbox) pgStatementsBuilder() *repo.Statements[*extdom.DataAuditOutbox, string] {
	return repo.NewStatements[*extdom.DataAuditOutbox, string]().
		WithStatement(repo.StatementFind, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxFind, dao.mapDefault)).
		WithStatement(repo.StatementList, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxList, dao.mapDefault)).
		WithStatement(repo.StatementCreate, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxCreate, dao.mapDefault)).
		WithStatement(repo.StatementChange, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxChange, dao.mapDefault)).
		WithStatement(repo.StatementDelete, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxDelete, dao.mapDefault)).
		WithStatement(statementGetTail, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxGetTail, dao.mapTail)).
		WithStatement(statementListPending, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxListPending, dao.mapDefault)).
		WithStatement(statementListProcessed, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxListProcessed, dao.mapDefault)).
		WithStatement(statementChangeStatus, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxChangeStatus, nil))
}
