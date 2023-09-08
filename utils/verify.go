package utils

var (
	IdVerify                         = Rules{"ID": []string{NotEmpty()}}
	ApiVerify                        = Rules{"Path": {NotEmpty()}, "Description": {NotEmpty()}, "ApiGroup": {NotEmpty()}, "Method": {NotEmpty()}}
	MenuVerify                       = Rules{"Path": {NotEmpty()}, "ParentId": {NotEmpty()}, "Name": {NotEmpty()}, "Component": {NotEmpty()}, "Sort": {Ge("0")}}
	MenuMetaVerify                   = Rules{"Title": {NotEmpty()}}
	LoginVerify                      = Rules{"CaptchaId": {NotEmpty()}, "Captcha": {NotEmpty()}, "Username": {NotEmpty()}, "Password": {NotEmpty()}}
	RegisterVerify                   = Rules{"Username": {NotEmpty()}, "NickName": {NotEmpty()}, "Password": {NotEmpty()}, "AuthorityId": {NotEmpty()}}
	PageInfoVerify                   = Rules{"Page": {NotEmpty()}, "PageSize": {NotEmpty()}}
	CustomerVerify                   = Rules{"CustomerName": {NotEmpty()}, "CustomerPhoneData": {NotEmpty()}}
	AutoCodeVerify                   = Rules{"Abbreviation": {NotEmpty()}, "StructName": {NotEmpty()}, "PackageName": {NotEmpty()}, "Fields": {NotEmpty()}}
	AutoPackageVerify                = Rules{"PackageName": {NotEmpty()}}
	AuthorityVerify                  = Rules{"AuthorityId": {NotEmpty()}, "AuthorityName": {NotEmpty()}}
	DepartmentRootVerify             = Rules{"Name": {NotEmpty()}, "ID": {NotEmpty()}}
	DepartmentVerify                 = Rules{"DepartmentName": {NotEmpty()}, "ParentDepartmentID": {NotEmpty()}} // 添加子部门的时候的参数，其实主要是对父部门做判断
	DepartmentIdVerify               = Rules{"ID": {NotEmpty()}}
	DepartmentSeqUpVerify            = Rules{"CurrentDepartmentID": {NotEmpty()}, "UpDepartmentID": {NotEmpty()}}
	DepartmentSeqDownVerify          = Rules{"CurrentDepartmentID": {NotEmpty()}, "DownDepartmentID": {NotEmpty()}}
	DepartmentIdAndNameVerfity       = Rules{"ID": {NotEmpty()}, "Name": {NotEmpty()}}
	DictionaryIdVerify               = Rules{"ID": {NotEmpty()}}
	DictionaryVerify                 = Rules{"Name": {NotEmpty()}, "Type": {NotEmpty()}, "Status": {NotEmpty()}, "Desc": {NotEmpty()}}                                        // 创建或更新字典时的参数非空校验
	DictionaryTypeVerify             = Rules{"Type": {NotEmpty()}}                                                                                                            // 字典英文名非空校验
	DataDictionaryDetailVerify       = Rules{"Label": {NotEmpty()}, "Value": {NotEmpty()}, "Status": {NotEmpty()}, "Sort": {NotEmpty()}, "SysDataDictionaryID": {NotEmpty()}} // 创建或更新字典项时的参数非空校验
	QuestionVerify                   = Rules{"QuestionBankID": {NotEmpty()}, "Level": {NotEmpty()}, "QuType": {NotEmpty()}, "Content": {NotEmpty()}}                          // 创建或更新试题时的参数非空校验
	QuestionBankVerify               = Rules{"QBName": {NotEmpty()}}                                                                                                          // 题库参数只需要名字非空校验
	AuthorityIdVerify                = Rules{"AuthorityId": {NotEmpty()}}
	OldAuthorityVerify               = Rules{"OldAuthorityId": {NotEmpty()}}
	ChangePasswordVerify             = Rules{"Password": {NotEmpty()}, "NewPassword": {NotEmpty()}}
	SetUserAuthorityVerify           = Rules{"AuthorityId": {NotEmpty()}}
	SetUserDepartmentAuthorityVerify = Rules{"AuthorityId": {NotEmpty()}, "DepartmentId": {NotEmpty()}}
	ExamCreatePaperVerify            = Rules{"name": {NotEmpty()}, "paperType": {}, "methGenPaper": {}, "questionIds": {NotEmpty()}, "eachScore": {NotEmpty()}}
	ExamUpdatePaperVerify            = Rules{"id": {NotEmpty()}, "name": {NotEmpty()}, "paperType": {}, "questionIds": {NotEmpty()}, "eachScore": {NotEmpty()}}
	ExamManagementVerify             = Rules{"exam_name": {NotEmpty()}, "exam_type": {NotEmpty()}, "exam_price": {NotEmpty()}, "passing_grade": {NotEmpty()}, "exam_duration": {NotEmpty()}, "number_of_exams": {NotEmpty()}}
	CreateMarkRecordVerify           = Rules{"examID":{NotEmpty()}}
)
