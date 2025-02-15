package types

import (
	"sync"

	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
)

type MaterializedStatementResultsIterator struct {
	isTableMode bool
	iterator    *ListElement[StatementResultRow]
	lock        sync.RWMutex
}

func (i *MaterializedStatementResultsIterator) HasReachedEnd() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()

	return i.iterator == nil
}

func (i *MaterializedStatementResultsIterator) GetNext() *StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Next()
	return row
}

func (i *MaterializedStatementResultsIterator) GetPrev() *StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Prev()
	return row
}

func (i *MaterializedStatementResultsIterator) Value() *StatementResultRow {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.iterator == nil || i.iterator.Value() == nil {
		return nil
	}

	row := i.iterator.Value()
	if !i.isTableMode {
		operationField := AtomicStatementResultField{
			Type:  "VARCHAR",
			Value: row.Operation.String(),
		}
		row.Fields = append([]StatementResultField{operationField}, row.Fields...)
	}
	return row
}

func (i *MaterializedStatementResultsIterator) Move(stepsToMove int) *StatementResultRow {
	for !i.HasReachedEnd() && stepsToMove != 0 {
		if stepsToMove < 0 {
			i.GetPrev()
			stepsToMove++
		} else {
			i.GetNext()
			stepsToMove--
		}
	}
	return i.Value()
}

type MaterializedStatementResults struct {
	isTableMode bool
	maxCapacity int
	headers     []string
	changelog   LinkedList[StatementResultRow]
	table       LinkedList[StatementResultRow]
	cache       map[string]*ListElement[StatementResultRow]
	lock        sync.RWMutex
}

func NewMaterializedStatementResults(headers []string, maxCapacity int) MaterializedStatementResults {
	return MaterializedStatementResults{
		isTableMode: true,
		maxCapacity: maxCapacity,
		headers:     headers,
		changelog:   NewLinkedList[StatementResultRow](),
		table:       NewLinkedList[StatementResultRow](),
		cache:       map[string]*ListElement[StatementResultRow]{},
	}
}

func (s *MaterializedStatementResults) GetTable() LinkedList[StatementResultRow] {
	return s.table
}

func (s *MaterializedStatementResults) Iterator(startFromBack bool) MaterializedStatementResultsIterator {
	s.lock.RLock()
	defer s.lock.RUnlock()

	list := s.table
	if !s.isTableMode {
		list = s.changelog
	}

	iterator := list.Front()
	if startFromBack {
		iterator = list.Back()
	}

	return MaterializedStatementResultsIterator{
		isTableMode: s.isTableMode,
		iterator:    iterator,
	}
}

func (s *MaterializedStatementResults) cleanup() {
	if s.changelog.Len() > s.maxCapacity {
		s.changelog.RemoveFront()
	}

	if s.table.Len() > s.maxCapacity {
		removedRow := s.table.RemoveFront()
		removedRowKey := removedRow.GetRowKey()
		delete(s.cache, removedRowKey)
	}
}

func (s *MaterializedStatementResults) Append(rows ...StatementResultRow) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	allValuesInserted := true
	for _, row := range rows {
		if len(row.Fields) != len(s.headers) {
			allValuesInserted = false
			continue
		}
		s.changelog.PushBack(row)

		rowKey := row.GetRowKey()
		if row.Operation.IsInsertOperation() {
			listPtr := s.table.PushBack(row)
			s.cache[rowKey] = listPtr
		} else {
			listPtr, ok := s.cache[rowKey]
			if ok {
				s.table.Remove(listPtr)
				delete(s.cache, rowKey)
			}
		}

		// if we are now over the capacity we need to remove some records
		s.cleanup()
	}
	return allValuesInserted
}

func (s *MaterializedStatementResults) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.isTableMode {
		return s.table.Len()
	}
	return s.changelog.Len()
}

func (s *MaterializedStatementResults) SetTableMode(isTableMode bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.isTableMode = isTableMode
}

func (s *MaterializedStatementResults) IsTableMode() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.isTableMode
}

func (s *MaterializedStatementResults) GetHeaders() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isTableMode {
		return s.headers
	}
	return append([]string{"Operation"}, s.headers...)
}

func (s *MaterializedStatementResults) ForEach(f func(rowIdx int, row *StatementResultRow)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	iterator := s.Iterator(false)
	rowIdx := 0
	for !iterator.HasReachedEnd() {
		f(rowIdx, iterator.GetNext())
		rowIdx++
	}
}

func (s *MaterializedStatementResults) GetMaxWidthPerColumn() []int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	columnWidths := make([]int, len(s.GetHeaders()))
	for colIdx, column := range s.GetHeaders() {
		columnWidths[colIdx] = max(utils.GetMaxStrWidth(column), columnWidths[colIdx])
	}

	s.ForEach(func(rowIdx int, row *StatementResultRow) {
		for colIdx, field := range row.Fields {
			columnWidths[colIdx] = max(utils.GetMaxStrWidth(field.ToString()), columnWidths[colIdx])
		}
	})
	return columnWidths
}

func (s *MaterializedStatementResults) GetMaxResults() int {
	return s.maxCapacity
}

func (s *MaterializedStatementResults) GetTableSize() int {
	return s.table.Len()
}

func (s *MaterializedStatementResults) GetChangelogSize() int {
	return s.changelog.Len()
}
