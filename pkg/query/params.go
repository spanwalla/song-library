// Package query предназначен для разбора параметров запроса, связанных с фильтрацией, сортировкой и пагинацией.
package query

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	AscendingSortOrder  = "asc"
	DescendingSortOrder = "desc"
)

// SortCriteria описывает критерий сортировки для одного поля.
type SortCriteria struct {
	Field string
	Order string
}

// Params хранит считанные параметры запроса.
type Params struct {
	values       url.Values
	Filters      map[string]string
	SortCriteria []SortCriteria
	Offset       int
	Limit        int
}

// NewParams создаёт экземпляр Params со значениями по умолчанию.
func NewParams(values url.Values) *Params {
	return &Params{
		values:       values,
		Filters:      make(map[string]string),
		SortCriteria: make([]SortCriteria, 0),
		Offset:       0,
		Limit:        0,
	}
}

// ParseFilters извлекает фильтры из values.
// Фильтры должны быть переданы в формате: filter[<поле>]=<значение> (например: &filter[name]=Muse&filter[group]=Cure).
func (p *Params) ParseFilters() {
	for key, vals := range p.values {
		if len(vals) > 0 && strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			field := key[len("filter[") : len(key)-1]
			p.Filters[field] = vals[0]
		}
	}
}

// ParseSortCriteria извлекает параметры сортировки из values.
// Ожидается формат: order_by=price:desc,name:asc,created_at (если не указать направление, по умолчанию будет asc).
func (p *Params) ParseSortCriteria() {
	sortByParam := p.values.Get("order_by")
	if len(sortByParam) > 0 {
		fields := strings.Split(sortByParam, ",")
		for _, field := range fields {
			parts := strings.Split(field, ":")
			field = parts[0]
			order := AscendingSortOrder
			if len(parts) > 1 && strings.ToLower(parts[1]) == DescendingSortOrder {
				order = DescendingSortOrder
			}
			p.SortCriteria = append(p.SortCriteria, SortCriteria{field, order})
		}
	}
}

// ParsePagination извлекает из values параметры пагинации (offset и limit).
func (p *Params) ParsePagination() {
	if o := p.values.Get("offset"); len(o) > 0 {
		if parsed, err := strconv.Atoi(o); err == nil {
			p.Offset = parsed
		}
	}

	if l := p.values.Get("limit"); len(l) > 0 {
		if parsed, err := strconv.Atoi(l); err == nil {
			p.Limit = parsed
		}
	}
}
