package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Query validation
type QueryValidationError struct {
    Parameter string `json:"parameter"`
    Value     string `json:"value"`
    Message   string `json:"message"`
}

type QueryValidator struct {
    paramPatterns map[string]*regexp.Regexp
    typeValidators map[string]func(string) bool
}

func NewQueryValidator() *QueryValidator {
    qv := &QueryValidator{
        paramPatterns: make(map[string]*regexp.Regexp),
        typeValidators: make(map[string]func(string) bool),
    }
    
    qv.AddParamPattern("default", `^[a-zA-Z][a-zA-Z0-9_]*$`)
    
    qv.typeValidators["number"] = func(v string) bool {
        matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, v)
        return matched
    }

    

    qv.typeValidators["boolean"] = func(v string) bool {
        v = strings.ToLower(v)
        return v == "true" || v == "false" || v == "1" || v == "0"
    }

    qv.typeValidators["date"] = func(v string) bool {
        matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, v)
        return matched
    }
    
    return qv
}

func (qv *QueryValidator) AddParamPattern(name, pattern string) error {
    regex, err := regexp.Compile(pattern)
    if err != nil {
        return fmt.Errorf("invalid pattern for %s: %v", name, err)
    }
    qv.paramPatterns[name] = regex
    return nil
}

func (qv *QueryValidator) AddTypeValidator(name string, validator func(string) bool) {
    qv.typeValidators[name] = validator
}

func (qv *QueryValidator) ValidateQuery(c fiber.Ctx, rules map[string]string) []QueryValidationError {
    var errors []QueryValidationError
    
    queries := c.Queries()
    
    for param, value := range queries {
        if !qv.validateParamName(param) {
            errors = append(errors, QueryValidationError{
                Parameter: param,
                Value:     value,
                Message:   "invalid parameter name format",
            })
            continue
        }
        
        expectedType, exists := rules[param]
        if !exists {
            errors = append(errors, QueryValidationError{
                Parameter: param,
                Value:     value,
                Message:   "unexpected parameter",
            })
            continue
        }
        
        if !qv.validateParamValue(value, expectedType) {
            errors = append(errors, QueryValidationError{
                Parameter: param,
                Value:     value,
                Message:   fmt.Sprintf("invalid value for type %s", expectedType),
            })
        }
    }
    
    return errors
}

func (qv *QueryValidator) validateParamName(param string) bool {
    pattern, exists := qv.paramPatterns["default"]
    if !exists {
        return true 
    }
    return pattern.MatchString(param)
}

func (qv *QueryValidator) validateParamValue(value, expectedType string) bool {
    validator, exists := qv.typeValidators[expectedType]
    if !exists {
        return true 
    }
    return validator(value)
}

func main() {

	fiberApp := fiber.New(fiber.Config{})

	fiberApp.Get("/:id", getAllUsersHandler)

}

func getAllUsersHandler(c fiber.Ctx) error {

	validator := NewQueryValidator()
    
    // Define validation rules
    rules := map[string]string{
        "age":    "number",
        "status": "string",
        "search": "string",
    }
    
    // Validate query parameters
    if errors := validator.ValidateQuery(c, rules); len(errors) > 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "errors": errors,
        })
    }

	return nil
}

