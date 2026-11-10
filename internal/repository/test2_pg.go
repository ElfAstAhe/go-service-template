package repository

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	pkgrepo "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
)

func (t2 *Test2) pgStatementsBuilder() *pkgrepo.Statements[*domain.Test2, string] {
	return pkgrepo.NewStatements[*domain.Test2, string]().
		WithStatement(pkgrepo.StatementFind, pkgrepo.NewStatement(postgres.SqlTest2Find, t2.mapDefault)).
		WithStatement(pkgrepo.StatementList, pkgrepo.NewStatement(postgres.SqlTest2List, t2.mapDefault)).
		WithStatement(pkgrepo.StatementCreate, pkgrepo.NewStatement(postgres.SqlTest2Create, t2.mapDefault)).
		WithStatement(pkgrepo.StatementChange, pkgrepo.NewStatement(postgres.SqlTest2Change, t2.mapDefault)).
		WithStatement(pkgrepo.StatementDelete, pkgrepo.NewStatement(postgres.SqlTest2Delete, t2.mapDefault)).
		WithStatement(statementTest2FindByCode, pkgrepo.NewStatement(postgres.SqlTest2FindByCode, t2.mapDefault))
}
