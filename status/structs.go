package status

import (
	"time"
)

const STATUS_SUCCESS_CODE = 0
const STATUS_SUCCESS_MSG = "SUCCESS"

const STATUS_RETRYING_CODE = 1
const STATUS_RETRYING_MSG = "RETRYING"

const STATUS_FAILED_CODE = 2
const STATUS_FAILED_MSG = "FAILED"

const STATUS_PENDING_CODE = 9
const STATUS_PENDING_MSG = "PENDING"

type ServiceStatus struct {
	Code    uint   `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`

	RetryTimes  uint      `json:"retry_times"`
	LastUpdated time.Time `json:"last_updated"`
}

func InitServiceStatus() ServiceStatus {
	status := ServiceStatus{
		RetryTimes: 0,
	}
	status.Pending()
	return status
}

func (status *ServiceStatus) SetMessage(message string) {
	status.Message = message
}

func (status *ServiceStatus) FlushMessage() {
	status.Message = ""
}

func (status *ServiceStatus) Success() {
	status.Code = STATUS_SUCCESS_CODE
	status.Status = STATUS_SUCCESS_MSG
	status.RetryTimes = 0
	status.LastUpdated = time.Now().UTC()
}

func (status *ServiceStatus) IsSuccess() bool {
	return status.Code == STATUS_SUCCESS_CODE
}

func (status *ServiceStatus) Retrying() {
	status.Code = STATUS_RETRYING_CODE
	status.Status = STATUS_RETRYING_MSG
	status.RetryTimes = status.RetryTimes + 1
	status.LastUpdated = time.Now().UTC()
}

func (status *ServiceStatus) IsRetrying() bool {
	return status.Code == STATUS_RETRYING_CODE
}

func (status *ServiceStatus) Failed() {
	status.Code = STATUS_FAILED_CODE
	status.Status = STATUS_FAILED_MSG
	status.RetryTimes = status.RetryTimes + 1
	status.LastUpdated = time.Now().UTC()
}

func (status *ServiceStatus) IsFailed() bool {
	return status.Code == STATUS_FAILED_CODE
}

func (status *ServiceStatus) Pending() {
	status.Code = STATUS_PENDING_CODE
	status.Status = STATUS_PENDING_MSG
	status.LastUpdated = time.Now().UTC()
}

func (status *ServiceStatus) IsPending() bool {
	return status.Code == STATUS_PENDING_CODE
}

var Status = InitServiceStatus()
