package apigateway

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Krunis/manager-o-order/packages/common"
)

func ValidateOrder(order *common.Order) error {
	var errorSlice []string

	if strings.TrimSpace(order.EmployeeID) == "" {
		log.Println("append employee_id is required")
		errorSlice = append(errorSlice, "employee_id is required")
	}
	if strings.TrimSpace(order.DepartmentID) == "" {
		log.Println("append department_id is required")
		errorSlice = append(errorSlice, "department_id is required")
	}

	for _, item := range order.Items {
		if err := ValidateItem(item); err != nil {
			errorSlice = append(errorSlice, fmt.Sprint(err))
		}
	}

	if order.Delivery == nil {
		log.Println("append delivery is required")
		errorSlice = append(errorSlice, "delivery is required")
	}

	if strings.TrimSpace(order.ConfirmationEmployeeID) == "" {
		log.Println("append confirmation_employee_id is required")
		errorSlice = append(errorSlice, "confirmation_employee_id is required")
	}
	if strings.TrimSpace(order.IdempotencyKey) == "" {
		log.Println("append idempotency_key is required")
		errorSlice = append(errorSlice, "idempotency_key is required")
	}

	if len(errorSlice) > 0 {
		return errors.New(strings.Join(errorSlice, ", "))
	}

	return nil
}

func ValidateItem(item *common.Item) error {
	var errorSlice []string

	if strings.TrimSpace(item.ID) == "" {
		log.Println("append item: id is required")
		errorSlice = append(errorSlice, "item: id is required")
	}
	if strings.TrimSpace(item.Name) == "" {
		log.Println("append item: name is required")
		errorSlice = append(errorSlice, "item: name is required")
	}
	if item.Count == 0 {
		log.Println("append item: count is required")
		errorSlice = append(errorSlice, "item: count is required")
	}
	if strings.TrimSpace(item.ConfirmationType) == "" {
		log.Println("append item: confirmation_type is required")
		errorSlice = append(errorSlice, "item: confirmation_type is required")
	}

	if len(errorSlice) > 0{
		return errors.New(strings.Join(errorSlice, ", "))
	}
	return nil
}

func ValidateEmployee(employee *common.Employee) error{
	var errorSlice []string

	if strings.TrimSpace(employee.ID) == ""{
		log.Println("append id is required")
		errorSlice = append(errorSlice, "id is required")
	}
	if strings.TrimSpace(employee.Department) == ""{
		log.Println("append department is required")
		errorSlice = append(errorSlice, "department is required")
	}

	if len(errorSlice) > 0{
		return errors.New(strings.Join(errorSlice, ","))
	}

	return nil
}