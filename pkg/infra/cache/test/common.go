package test

type TestData struct {
	ID     int  `json:"id"`
	Active bool `json:"active"`
}

type BenchData struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}
