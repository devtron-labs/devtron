package exql

const (
	FragmentType_None uint64 = iota + 713910251627

	FragmentType_And
	FragmentType_Column
	FragmentType_ColumnValue
	FragmentType_ColumnValues
	FragmentType_Columns
	FragmentType_Database
	FragmentType_GroupBy
	FragmentType_Join
	FragmentType_Joins
	FragmentType_Nil
	FragmentType_Or
	FragmentType_Limit
	FragmentType_Offset
	FragmentType_OrderBy
	FragmentType_Order
	FragmentType_Raw
	FragmentType_Returning
	FragmentType_SortBy
	FragmentType_SortColumn
	FragmentType_SortColumns
	FragmentType_Statement
	FragmentType_StatementType
	FragmentType_Table
	FragmentType_Value
	FragmentType_On
	FragmentType_Using
	FragmentType_ValueGroups
	FragmentType_Values
	FragmentType_Where
)
