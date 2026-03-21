package apigateway

import (
	"errors"
	"strings"

	"github.com/Krunis/manager-o-order/packages/common"
)

func ValidateOrder(order *common.Order) error {
	var errorSlice []string

	if strings.TrimSpace(order.EmployeeID) == ""{
		errorSlice = append(errorSlice, "employee_id is required")
	}
	if strings.TrimSpace(order.DepartmentID) == ""{
		errorSlice = append(errorSlice, "department_id is required")
	}

	for _, item := range order.Items{
		errorSlice = append(errorSlice, ValidateItem(item))
	}

	if order.Delivery == nil{
		errorSlice = append(errorSlice, "delivery is required")
	}

	if strings.TrimSpace(order.ConfirmationEmployeeID) == ""{
		errorSlice = append(errorSlice, "confirmation_employee_id is required")
	}
	if strings.TrimSpace(order.IdempotencyKey) == ""{
		errorSlice = append(errorSlice, "idempotency_key is required")
	}
	
	if len(errorSlice) > 0 {
		return errors.New(strings.Join(errorSlice, ", "))
	}

	return nil
}

func ValidateItem(item *common.Item) string {
	var errorSlice []string

	if strings.TrimSpace(item.ID) == ""{
		errorSlice = append(errorSlice, "item: id is required")
	}
	if strings.TrimSpace(item.Name) == ""{
		errorSlice = append(errorSlice, "item: name is required")
	}
	if item.Count == 0{
		errorSlice = append(errorSlice, "item: count is required")
	}
	if strings.TrimSpace(item.ConfirmationType) == ""{
		errorSlice = append(errorSlice, "item: confirmation_type is required")
	}

	return strings.Join(errorSlice, ", ")
}
