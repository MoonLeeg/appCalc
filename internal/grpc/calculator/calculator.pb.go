
package calculator

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetTaskRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	AgentId       string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTaskRequest) Reset() {
	*x = GetTaskRequest{}
	mi := &file_calculator_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskRequest) ProtoMessage() {}

func (x *GetTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetTaskRequest) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{0}
}

func (x *GetTaskRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

type GetTaskResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	TaskInfo      isGetTaskResponse_TaskInfo `protobuf_oneof:"task_info"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTaskResponse) Reset() {
	*x = GetTaskResponse{}
	mi := &file_calculator_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskResponse) ProtoMessage() {}

func (x *GetTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetTaskResponse) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{1}
}

func (x *GetTaskResponse) GetTaskInfo() isGetTaskResponse_TaskInfo {
	if x != nil {
		return x.TaskInfo
	}
	return nil
}

func (x *GetTaskResponse) GetTask() *Task {
	if x != nil {
		if x, ok := x.TaskInfo.(*GetTaskResponse_Task); ok {
			return x.Task
		}
	}
	return nil
}

func (x *GetTaskResponse) GetNoTask() *NoTaskAvailable {
	if x != nil {
		if x, ok := x.TaskInfo.(*GetTaskResponse_NoTask); ok {
			return x.NoTask
		}
	}
	return nil
}

type isGetTaskResponse_TaskInfo interface {
	isGetTaskResponse_TaskInfo()
}

type GetTaskResponse_Task struct {
	Task *Task `protobuf:"bytes,1,opt,name=task,proto3,oneof"` // Задача для выполнения
}

type GetTaskResponse_NoTask struct {
	NoTask *NoTaskAvailable `protobuf:"bytes,2,opt,name=no_task,json=noTask,proto3,oneof"` // Сообщение, что задач нет
}

func (*GetTaskResponse_Task) isGetTaskResponse_TaskInfo() {}

func (*GetTaskResponse_NoTask) isGetTaskResponse_TaskInfo() {}

type Task struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Id              int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`                                                    // ID задачи в БД
	Arg1            float64                `protobuf:"fixed64,2,opt,name=arg1,proto3" json:"arg1,omitempty"`                                               // Первый аргумент
	Arg2            float64                `protobuf:"fixed64,3,opt,name=arg2,proto3" json:"arg2,omitempty"`                                               // Второй аргумент
	Operation       string                 `protobuf:"bytes,4,opt,name=operation,proto3" json:"operation,omitempty"`                                       // Операция (+, -, *, /)
	OperationTimeMs int32                  `protobuf:"varint,5,opt,name=operation_time_ms,json=operationTimeMs,proto3" json:"operation_time_ms,omitempty"` // Время выполнения в мс
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *Task) Reset() {
	*x = Task{}
	mi := &file_calculator_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Task) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{2}
}

func (x *Task) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Task) GetArg1() float64 {
	if x != nil {
		return x.Arg1
	}
	return 0
}

func (x *Task) GetArg2() float64 {
	if x != nil {
		return x.Arg2
	}
	return 0
}

func (x *Task) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *Task) GetOperationTimeMs() int32 {
	if x != nil {
		return x.OperationTimeMs
	}
	return 0
}

type NoTaskAvailable struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	RetryAfterSeconds int32 `protobuf:"varint,1,opt,name=retry_after_seconds,json=retryAfterSeconds,proto3" json:"retry_after_seconds,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *NoTaskAvailable) Reset() {
	*x = NoTaskAvailable{}
	mi := &file_calculator_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NoTaskAvailable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NoTaskAvailable) ProtoMessage() {}

func (x *NoTaskAvailable) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*NoTaskAvailable) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{3}
}

func (x *NoTaskAvailable) GetRetryAfterSeconds() int32 {
	if x != nil {
		return x.RetryAfterSeconds
	}
	return 0
}

type SubmitResultRequest struct {
	state  protoimpl.MessageState `protogen:"open.v1"`
	TaskId int64                  `protobuf:"varint,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"` // ID выполненной задачи
	ResultStatus  isSubmitResultRequest_ResultStatus `protobuf_oneof:"result_status"`
	AgentId       string                             `protobuf:"bytes,4,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"` // ID агента, выполнившего задачу
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitResultRequest) Reset() {
	*x = SubmitResultRequest{}
	mi := &file_calculator_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitResultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitResultRequest) ProtoMessage() {}

func (x *SubmitResultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SubmitResultRequest) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{4}
}

func (x *SubmitResultRequest) GetTaskId() int64 {
	if x != nil {
		return x.TaskId
	}
	return 0
}

func (x *SubmitResultRequest) GetResultStatus() isSubmitResultRequest_ResultStatus {
	if x != nil {
		return x.ResultStatus
	}
	return nil
}

func (x *SubmitResultRequest) GetResult() float64 {
	if x != nil {
		if x, ok := x.ResultStatus.(*SubmitResultRequest_Result); ok {
			return x.Result
		}
	}
	return 0
}

func (x *SubmitResultRequest) GetError() *TaskError {
	if x != nil {
		if x, ok := x.ResultStatus.(*SubmitResultRequest_Error); ok {
			return x.Error
		}
	}
	return nil
}

func (x *SubmitResultRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

type isSubmitResultRequest_ResultStatus interface {
	isSubmitResultRequest_ResultStatus()
}

type SubmitResultRequest_Result struct {
	Result float64 `protobuf:"fixed64,2,opt,name=result,proto3,oneof"` // Успешный результат вычисления
}

type SubmitResultRequest_Error struct {
	Error *TaskError `protobuf:"bytes,3,opt,name=error,proto3,oneof"` // Информация об ошибке
}

func (*SubmitResultRequest_Result) isSubmitResultRequest_ResultStatus() {}

func (*SubmitResultRequest_Error) isSubmitResultRequest_ResultStatus() {}

type TaskError struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"` // Сообщение об ошибке (например, "деление на ноль")
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskError) Reset() {
	*x = TaskError{}
	mi := &file_calculator_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskError) ProtoMessage() {}

func (x *TaskError) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TaskError) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{5}
}

func (x *TaskError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SubmitResultResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Acknowledged  bool                   `protobuf:"varint,1,opt,name=acknowledged,proto3" json:"acknowledged,omitempty"` // Подтверждение получения результата
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitResultResponse) Reset() {
	*x = SubmitResultResponse{}
	mi := &file_calculator_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitResultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitResultResponse) ProtoMessage() {}

func (x *SubmitResultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_calculator_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SubmitResultResponse) Descriptor() ([]byte, []int) {
	return file_calculator_proto_rawDescGZIP(), []int{6}
}

func (x *SubmitResultResponse) GetAcknowledged() bool {
	if x != nil {
		return x.Acknowledged
	}
	return false
}

var File_calculator_proto protoreflect.FileDescriptor

const file_calculator_proto_rawDesc = "" +
	"\n" +
	"\x10calculator.proto\x12\n" +
	"calculator\"+\n" +
	"\x0eGetTaskRequest\x12\x19\n" +
	"\bagent_id\x18\x01 \x01(\tR\aagentId\"~\n" +
	"\x0fGetTaskResponse\x12&\n" +
	"\x04task\x18\x01 \x01(\v2\x10.calculator.TaskH\x00R\x04task\x126\n" +
	"\ano_task\x18\x02 \x01(\v2\x1b.calculator.NoTaskAvailableH\x00R\x06noTaskB\v\n" +
	"\ttask_info\"\x88\x01\n" +
	"\x04Task\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x12\n" +
	"\x04arg1\x18\x02 \x01(\x01R\x04arg1\x12\x12\n" +
	"\x04arg2\x18\x03 \x01(\x01R\x04arg2\x12\x1c\n" +
	"\toperation\x18\x04 \x01(\tR\toperation\x12*\n" +
	"\x11operation_time_ms\x18\x05 \x01(\x05R\x0foperationTimeMs\"A\n" +
	"\x0fNoTaskAvailable\x12.\n" +
	"\x13retry_after_seconds\x18\x01 \x01(\x05R\x11retryAfterSeconds\"\xa3\x01\n" +
	"\x13SubmitResultRequest\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\x03R\x06taskId\x12\x18\n" +
	"\x06result\x18\x02 \x01(\x01H\x00R\x06result\x12-\n" +
	"\x05error\x18\x03 \x01(\v2\x15.calculator.TaskErrorH\x00R\x05error\x12\x19\n" +
	"\bagent_id\x18\x04 \x01(\tR\aagentIdB\x0f\n" +
	"\rresult_status\"%\n" +
	"\tTaskError\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\":\n" +
	"\x14SubmitResultResponse\x12\"\n" +
	"\facknowledged\x18\x01 \x01(\bR\facknowledged2\xaf\x01\n" +
	"\x16CalculatorAgentService\x12B\n" +
	"\aGetTask\x12\x1a.calculator.GetTaskRequest\x1a\x1b.calculator.GetTaskResponse\x12Q\n" +
	"\fSubmitResult\x12\x1f.calculator.SubmitResultRequest\x1a .calculator.SubmitResultResponseB Z\x1ecalculator/pkg/grpc/calculatorb\x06proto3"

var (
	file_calculator_proto_rawDescOnce sync.Once
	file_calculator_proto_rawDescData []byte
)

func file_calculator_proto_rawDescGZIP() []byte {
	file_calculator_proto_rawDescOnce.Do(func() {
		file_calculator_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_calculator_proto_rawDesc), len(file_calculator_proto_rawDesc)))
	})
	return file_calculator_proto_rawDescData
}

var file_calculator_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_calculator_proto_goTypes = []any{
	(*GetTaskRequest)(nil),       // 0: calculator.GetTaskRequest
	(*GetTaskResponse)(nil),      // 1: calculator.GetTaskResponse
	(*Task)(nil),                 // 2: calculator.Task
	(*NoTaskAvailable)(nil),      // 3: calculator.NoTaskAvailable
	(*SubmitResultRequest)(nil),  // 4: calculator.SubmitResultRequest
	(*TaskError)(nil),            // 5: calculator.TaskError
	(*SubmitResultResponse)(nil), // 6: calculator.SubmitResultResponse
}
var file_calculator_proto_depIdxs = []int32{
	2, // 0: calculator.GetTaskResponse.task:type_name -> calculator.Task
	3, // 1: calculator.GetTaskResponse.no_task:type_name -> calculator.NoTaskAvailable
	5, // 2: calculator.SubmitResultRequest.error:type_name -> calculator.TaskError
	0, // 3: calculator.CalculatorAgentService.GetTask:input_type -> calculator.GetTaskRequest
	4, // 4: calculator.CalculatorAgentService.SubmitResult:input_type -> calculator.SubmitResultRequest
	1, // 5: calculator.CalculatorAgentService.GetTask:output_type -> calculator.GetTaskResponse
	6, // 6: calculator.CalculatorAgentService.SubmitResult:output_type -> calculator.SubmitResultResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_calculator_proto_init() }
func file_calculator_proto_init() {
	if File_calculator_proto != nil {
		return
	}
	file_calculator_proto_msgTypes[1].OneofWrappers = []any{
		(*GetTaskResponse_Task)(nil),
		(*GetTaskResponse_NoTask)(nil),
	}
	file_calculator_proto_msgTypes[4].OneofWrappers = []any{
		(*SubmitResultRequest_Result)(nil),
		(*SubmitResultRequest_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_calculator_proto_rawDesc), len(file_calculator_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_calculator_proto_goTypes,
		DependencyIndexes: file_calculator_proto_depIdxs,
		MessageInfos:      file_calculator_proto_msgTypes,
	}.Build()
	File_calculator_proto = out.File
	file_calculator_proto_goTypes = nil
	file_calculator_proto_depIdxs = nil
}