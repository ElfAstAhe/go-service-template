package repository

type BaseOwnedQueryBuilders struct {
	findBuilder            QueryBuilderFunc
	listBuilder            QueryBuilderFunc
	listAllBuilder         QueryBuilderFunc
	listAllByOwnersBuilder QueryBuilderFunc
	createBuilder          QueryBuilderFunc
	changeBuilder          QueryBuilderFunc
	deleteAllBuilder       QueryBuilderFunc
	deleteBuilder          QueryBuilderFunc
}

func newBaseOwnerQueryBuilders() *BaseOwnedQueryBuilders {
	return &BaseOwnedQueryBuilders{}
}

func (bq *BaseOwnedQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

func (bq *BaseOwnedQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

func (bq *BaseOwnedQueryBuilders) GetListAll() QueryBuilderFunc {
	return bq.listAllBuilder
}

func (bq *BaseOwnedQueryBuilders) GetListAllByOwners() QueryBuilderFunc {
	return bq.listAllByOwnersBuilder
}

func (bq *BaseOwnedQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

func (bq *BaseOwnedQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

func (bq *BaseOwnedQueryBuilders) DeleteAll() QueryBuilderFunc {
	return bq.deleteAllBuilder
}

func (bq *BaseOwnedQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

type BaseOwnedQueryBuildersBuilder struct {
	instance *BaseOwnedQueryBuilders
}

func NewBaseOwnedQueryBuildersBuilder() *BaseOwnedQueryBuildersBuilder {
	return &BaseOwnedQueryBuildersBuilder{
		instance: &BaseOwnedQueryBuilders{},
	}
}

func (bbo *BaseOwnedQueryBuildersBuilder) NewInstance() *BaseOwnedQueryBuildersBuilder {
	bbo.instance = newBaseOwnerQueryBuilders()

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.findBuilder = findBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listBuilder = listBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithListAll(listAllBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllBuilder = listAllBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithListAllByOwners(listAllByOwnersBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllByOwnersBuilder = listAllByOwnersBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.createBuilder = createBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.changeBuilder = change

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithDelete(deleteBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.deleteBuilder = deleteBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) Build() *BaseOwnedQueryBuilders {
	return bbo.instance
}
