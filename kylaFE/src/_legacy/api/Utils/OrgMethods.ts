import { RpcError } from "grpc-web";
import {
	appServiceClient,
	authServiceClient,
	branchServiceClient,
	contactServiceClient,
	departmentServiceClient,
	groupServiceClient,
	invitationServiceClient,
	leaveServiceClient,
	orgServiceClient,
	permissionServiceClient,
	roleServiceClient,
	scheduleBreakServiceClient,
	shiftScheduleServiceClient,
	shiftServiceClient,
	teamServiceClient,
	userServiceClient
} from "../globalClient/GlobalClients";
import {
	ChangePasswordRequest,
	ChangePasswordResponse,
	CheckUserPermissionRequest,
	CheckUserPermissionResponse,
	ForgotPasswordRequest,
	ForgotPasswordResponse,
	LoginRequest,
	LoginResponse,
	ReadAuthContextRequest,
	ReadAuthContextResponse,
	RefreshTokenRequest,
	RefreshTokenResponse,
} from "../../pb/auth";
import { AuthServiceClient } from "../../pb/auth.client";
import {
	CreateDefaultStructureRequest,
	CreateDefaultStructureResponse,
	CreateOrganisationRequest,
	CreateOrganisationResponse,
	CreateStructureWithTemplateRequest,
	CreateStructureWithTemplateResponse,
	DeleteOrganisationRequest,
	DeleteOrganisationResponse,
	Organisation,
	ReadMeRequest,
	ReadMeResponse,
	ReadOrganisationRequest,
	ReadOrganisationResponse,
	ReadOrganisationsRequest,
	ReadOrganisationsResponse,
	ReadTemplatesRequest,
	ReadTemplatesResponse,
	UpdateOrganisationRequest,
	UpdateOrganisationResponse,
} from "../../pb/organisations";
import { OrganisationServiceClient } from "../../pb/organisations.client";
import {
	ActivateUserAccountRequest,
	ActivateUserAccountResponse,
	CreateUserRequest,
	CreateUserResponse,
	DeactivateUserRequest,
	DeactivateUserResponse,
	DeleteUserRequest,
	DeleteUserResponse,
	ReadUserRequest,
	ReadUserResponse,
	ReadUsersRequest,
	ReadUsersResponse,
	RoleToUserRequest,
	RoleToUserResponse,
	SearchUsersRequest,
	SearchUsersResponse,
	SignUpRequest,
	SignUpResponse,
	UpdateUserRequest,
	UpdateUserResponse,
	UsersToRoleRequest,
	UsersToRoleResponse,
} from "../../pb/user";
import { UserServiceClient } from "../../pb/user.client";
import { makeGRPCCall, tokenObject } from "./rpcUtils";

import {
	ApproveAppRequest,
	ApproveAppResponse,
	CreateAppRequest,
	CreateAppResponse,
	CreateAppWithTemplateRequest,
	CreateAppWithTemplateResponse,
	CreateTemplateAppRequest,
	CreateTemplateAppResponse,
	DeleteAppRequest,
	DeleteAppResponse,
	ReadAppRequest,
	ReadAppResponse,
	ReadAppsRequest,
	ReadAppsResponse,
	ReadAppTemplatesRequest,
	ReadAppTemplatesResponse,
	UpdateAppRequest,
	UpdateAppResponse,
} from "../../pb/apps";
import { AppServiceClient } from "../../pb/apps.client";
import { BranchServiceClient } from "../../pb/branch.client";
import { ContactServiceClient } from "../../pb/contact.client";
import {
	CreateGroupRequest,
	CreateGroupResponse,
	DeleteGroupRequest,
	DeleteGroupResponse,
	ReadGroupContactsRequest,
	ReadGroupContactsResponse,
	ReadGroupRequest,
	ReadGroupResponse,
	ReadGroupsRequest,
	ReadGroupsResponse,
	UpdateGroupRequest,
	UpdateGroupResponse,
} from "../../pb/contact_groups";
import { GroupServiceClient } from "../../pb/contact_groups.client";
import {
	CreateDepartmentRequest,
	CreateDepartmentResponse,
	DeleteDepartmentRequest,
	DeleteDepartmentResponse,
	ReadBranchDepartmentsRequest,
	ReadBranchDepartmentsResponse,
	ReadDepartmentRequest,
	ReadDepartmentResponse,
	ReadDepartmentsRequest,
	ReadDepartmentsResponse,
	UpdateDepartmentRequest,
	UpdateDepartmentResponse,
} from "../../pb/department";
import { DepartmentServiceClient } from "../../pb/department.client";
import {
	AcceptInviteRequest,
	CancelInviteRequest,
	CreateInviteRequest,
	GetInviteRequest,
	Invitation,
	ListInvitationsRequest,
	ListInvitationsResponse,
	RejectInviteRequest,
} from "../../pb/invitation";
import { InvitationServiceClient } from "../../pb/invitation.client";
import {
	AppealLeaveRequestRequest,
	AppealLeaveRequestResponse,
	ApproveLeaveRequestRequest,
	ApproveLeaveRequestResponse,
	CancelLeaveRequestRequest,
	CancelLeaveRequestResponse,
	CreateLeaveRequestRequest,
	CreateLeaveRequestResponse,
	CreateLeaveTypeRequest,
	CreateLeaveTypeResponse,
	DeleteLeaveTypeRequest,
	DeleteLeaveTypeResponse,
	EndLeaveRequestRequest,
	EndLeaveRequestResponse,
	ReadLeaveBalanceRequest,
	ReadLeaveBalanceResponse,
	ReadLeaveRequestRequest,
	ReadLeaveRequestResponse,
	ReadLeaveRequestsMetricsRequest,
	ReadLeaveRequestsMetricsResponse,
	ReadLeaveRequestsRequest,
	ReadLeaveRequestsResponse,
	ReadLeaveTypeRequest,
	ReadLeaveTypeResponse,
	ReadLeaveTypesRequest,
	ReadLeaveTypesResponse,
	ReadMyLeaveRequestsRequest,
	ReadMyLeaveRequestsResponse,
	RejectLeaveRequestRequest,
	RejectLeaveRequestResponse,
	UpdateLeaveTypeRequest,
	UpdateLeaveTypeResponse,
} from "../../pb/leave";
import { LeaveServiceClient } from "../../pb/leave.client";
import {
	CreateRoleRequest,
	CreateRoleResponse,
	DeleteRoleRequest,
	DeleteRoleResponse,
	PermissionToRoleRequest,
	PermissionToRoleResponse,
	ReadPermissionsRequest,
	ReadPermissionsResponse,
	ReadRoleRequest,
	ReadRoleResponse,
	//ReadRolesInOrganisationRequest,
	ReadRolesRequest,
	ReadRolesResponse,
	UpdateRoleRequest,
	UpdateRoleResponse,
} from "../../pb/rbac";
import { PermissionServiceClient, RoleServiceClient } from "../../pb/rbac.client";
import {
	BreakServiceClient,
	ShiftScheduleServiceClient,
	ShiftServiceClient,
} from "../../pb/shifts.client";
import {
	CreateTeamRequest,
	CreateTeamResponse,
	DeleteTeamRequest,
	DeleteTeamResponse,
	ReadTeamListRequest,
	ReadTeamListResponse,
	ReadTeamRequest,
	ReadTeamResponse,
	UpdateTeamRequest,
	UpdateTeamResponse,
} from "../../pb/team";
import { TeamServiceClient } from "../../pb/team.client";
import type { CreateBranchRequest, CreateBranchResponse, ReadBranchRequest, ReadBranchResponse, ReadBranchesRequest, ReadBranchesResponse, UpdateBranchRequest, UpdateBranchResponse, DeleteBranchRequest, DeleteBranchResponse } from "@/pb/branch";
import type { ReadContactsRequest, ReadContactsResponse, ReadContactRequest, ReadContactResponse, CreateContactRequest, CreateContactResponse, UpdateContactRequest, UpdateContactResponse, DeleteContactRequest, DeleteContactResponse, BulkContactsImportRequest, BulkContactsImportResponse, BulkContactsExportRequest, BulkContactsExportResponse, ReadCustomFieldDefinitionsRequest, ReadCustomFieldDefinitionsResponse, FindContactDuplicatesRequest, FindContactDuplicatesResponse, MergeDuplicateContactsRequest, MergeDuplicateContactsResponse } from "@/pb/contact";
import type { CreateShiftRequest, CreateShiftResponse, ReadShiftRequest, ReadShiftResponse, ListShiftsRequest, ListShiftsResponse, UpdateShiftRequest, UpdateShiftResponse, DeleteShiftRequest, DeleteShiftResponse, GetUserShiftsRequest, GetUserShiftsResponse, AssignUsersToShiftRequest, AssignUsersToShiftResponse, RemoveUsersFromShiftRequest, RemoveUsersFromShiftResponse, ListShiftUsersRequest, ListShiftUsersResponse, ReadScheduleRequest, ReadScheduleResponse, UpdateScheduleRequest, UpdateScheduleResponse, DeleteScheduleRequest, DeleteScheduleResponse, ListSchedulesRequest, ListSchedulesResponse, ClockInRequest, ClockInResponse, ClockOutRequest, ClockOutResponse, CreateBreakRequest, CreateBreakResponse, ReadBreakRequest, ReadBreakResponse, UpdateBreakRequest, UpdateBreakResponse, DeleteBreakRequest, DeleteBreakResponse, ListBreaksRequest, ListBreaksResponse, TakeBreakRequest, TakeBreakResponse, ResumeBreakRequest, ResumeBreakResponse } from "@/pb/shifts";

export const userLogin = (data: LoginRequest): Promise<LoginResponse> => {
	return new Promise<LoginResponse>((resolve, reject) => {
		makeGRPCCall<LoginRequest, AuthServiceClient, LoginResponse>(
			data,
			authServiceClient,
			"login"
		)
			.then((response: LoginResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const RefreshToken = (
	data: RefreshTokenRequest
): Promise<RefreshTokenResponse> => {
	return new Promise<RefreshTokenResponse>((resolve, reject) => {
		makeGRPCCall<RefreshTokenRequest, AuthServiceClient, RefreshTokenResponse>(
			data,
			authServiceClient,
			"refreshToken"
		)
			.then((response: RefreshTokenResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const ReadMe = (data: ReadMeRequest): Promise<ReadMeResponse> => {
	return new Promise<ReadMeResponse>((resolve, reject) => {
		makeGRPCCall<ReadMeRequest, OrganisationServiceClient, ReadMeResponse>(
			data,
			orgServiceClient,
			"readMe"
		)
			.then((res: ReadMeResponse) => {
				resolve(res);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// forgot password
export const _ForgotPassword = (data: ForgotPasswordRequest) => {
	return new Promise<ForgotPasswordResponse>((resolve, reject) => {
		makeGRPCCall<
			ForgotPasswordRequest,
			AuthServiceClient,
			ForgotPasswordResponse
		>(data, authServiceClient, "forgotPassword")
			.then((response: ForgotPasswordResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// change password

export const ChangePassword = (data: ChangePasswordRequest) => {
	return new Promise<ChangePasswordResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangePasswordRequest,
			AuthServiceClient,
			ChangePasswordResponse
		>(data, authServiceClient, "changePassword")
			.then((response: ChangePasswordResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// activate user account

export const ActivateUserAccount = (data: ActivateUserAccountRequest) => {
	return new Promise<ActivateUserAccountResponse>((resolve, reject) => {
		makeGRPCCall<
			ActivateUserAccountRequest,
			UserServiceClient,
			ActivateUserAccountResponse
		>(data, userServiceClient, "activateUserAccount")
			.then((response: ActivateUserAccountResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const CheckPermissions = (data: CheckUserPermissionRequest) => {
	return new Promise<CheckUserPermissionResponse>((resolve, reject) => {
		makeGRPCCall<
			CheckUserPermissionRequest,
			AuthServiceClient,
			CheckUserPermissionResponse
		>(data, authServiceClient, "checkUserPermission")
			.then((response: CheckUserPermissionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// Create an organisation

export const CreateOrganisation = (data: Organisation) => {
	const orgData: CreateOrganisationRequest = {
		organisation: data,
	};
	return new Promise<CreateOrganisationResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateOrganisationRequest,
			OrganisationServiceClient,
			CreateOrganisationResponse
		>(orgData, orgServiceClient, "createOrganisation")
			.then((response: CreateOrganisationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readOrganisation = (
	data: ReadOrganisationRequest
): Promise<ReadOrganisationResponse> => {
	return new Promise<ReadOrganisationResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadOrganisationsRequest,
			OrganisationServiceClient,
			ReadOrganisationResponse
		>(data, orgServiceClient, "readOrganisation")
			.then((response: ReadOrganisationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readOrganisations = (
	data: ReadOrganisationsRequest
): Promise<ReadOrganisationsResponse> => {
	return new Promise<ReadOrganisationsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadOrganisationsRequest,
			OrganisationServiceClient,
			ReadOrganisationsResponse
		>(data, orgServiceClient, "readOrganisations")
			.then((response: ReadOrganisationsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const UpdateOrganisation = (
	data: UpdateOrganisationRequest
): Promise<UpdateOrganisationResponse> => {
	return new Promise<UpdateOrganisationResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateOrganisationRequest,
			OrganisationServiceClient,
			UpdateOrganisationResponse
		>(data, orgServiceClient, "updateOrganisation")
			.then((response: UpdateOrganisationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteOrganisation = (
	data: DeleteOrganisationRequest
): Promise<DeleteOrganisationResponse> => {
	return new Promise<DeleteOrganisationResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteOrganisationRequest,
			OrganisationServiceClient,
			DeleteOrganisationResponse
		>(data, orgServiceClient, "deleteOrganization")
			.then((response: DeleteOrganisationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTemplates = (
	data: ReadTemplatesRequest
): Promise<ReadTemplatesResponse> => {
	return new Promise<ReadTemplatesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTemplatesRequest,
			OrganisationServiceClient,
			ReadTemplatesResponse
		>(data, orgServiceClient, "readTemplates")
			.then((response: ReadTemplatesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createDefaultStructure = (
	data: CreateDefaultStructureRequest
): Promise<CreateDefaultStructureResponse> => {
	return new Promise<CreateDefaultStructureResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateDefaultStructureRequest,
			OrganisationServiceClient,
			CreateDefaultStructureResponse
		>(data, orgServiceClient, "createDefaultStructure")
			.then((response: CreateDefaultStructureResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createStructureWithTemplate = (
	data: CreateStructureWithTemplateRequest
): Promise<CreateStructureWithTemplateResponse> => {
	return new Promise<CreateStructureWithTemplateResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateStructureWithTemplateRequest,
			OrganisationServiceClient,
			CreateStructureWithTemplateResponse
		>(data, orgServiceClient, "createStructureWithTemplate")
			.then((response: CreateStructureWithTemplateResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

//WIP: Confirm Create User
export const createUser = (
	data: CreateUserRequest
): Promise<CreateUserResponse> => {
	return new Promise<CreateUserResponse>((resolve, reject) => {
		makeGRPCCall<CreateUserRequest, UserServiceClient, CreateUserResponse>(
			data,
			userServiceClient,
			"createUser"
		)
			.then((response: CreateUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readUser = (data: ReadUserRequest): Promise<ReadUserResponse> => {
	return new Promise<ReadUserResponse>((resolve, reject) => {
		makeGRPCCall<ReadUserRequest, UserServiceClient, ReadUserResponse>(
			data,
			userServiceClient,
			"readUser"
		)
			.then((response: ReadUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const updateUser = (
	data: UpdateUserRequest
): Promise<UpdateUserResponse> => {
	return new Promise<UpdateUserResponse>((resolve, reject) => {
		makeGRPCCall<UpdateUserRequest, UserServiceClient, UpdateUserResponse>(
			data,
			userServiceClient,
			"updateUser"
		)
			.then((response: UpdateUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deactivateUser = (
	data: DeactivateUserRequest
): Promise<DeactivateUserResponse> => {
	return new Promise<DeactivateUserResponse>((resolve, reject) => {
		makeGRPCCall<
			DeactivateUserRequest,
			UserServiceClient,
			DeactivateUserResponse
		>(data, userServiceClient, "deactivateUser")
			.then((response: DeactivateUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteUser = (
	data: DeleteUserRequest
): Promise<DeleteUserResponse> => {
	return new Promise<DeleteUserResponse>((resolve, reject) => {
		makeGRPCCall<DeleteUserRequest, UserServiceClient, DeleteUserResponse>(
			data,
			userServiceClient,
			"deleteUser"
		)
			.then((response: DeleteUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const signUp = (data: SignUpRequest): Promise<SignUpResponse> => {
	return new Promise<SignUpResponse>((resolve, reject) => {
		makeGRPCCall<SignUpRequest, UserServiceClient, SignUpResponse>(
			data,
			userServiceClient,
			"signUp"
		)
			.then((response: SignUpResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// MARK: User invitations.
// invite a User
export const inviteUser = (data: CreateInviteRequest): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<CreateInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"createInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// CreateInvitation
// GetInvitation
// ListInvitations
// UpdateInvitation
// DeleteInvitation
// AcceptInvitation
// RejectInvitation
// CancelInvitation

export const getInvitation = (data: GetInviteRequest): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<GetInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"getInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const listInvitations = (
	data: ListInvitationsRequest
): Promise<ListInvitationsResponse> => {
	return new Promise<ListInvitationsResponse>((resolve, reject) => {
		makeGRPCCall<
			ListInvitationsRequest,
			InvitationServiceClient,
			ListInvitationsResponse
		>(data, invitationServiceClient, "listInvitations")
			.then((response: ListInvitationsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const acceptInvitation = (
	data: AcceptInviteRequest
): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<AcceptInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"acceptInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const rejectInvitation = (
	data: RejectInviteRequest
): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<RejectInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"rejectInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const cancelInvitation = (
	data: CancelInviteRequest
): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<CancelInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"cancelInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteInvitation = (
	data: CancelInviteRequest
): Promise<Invitation> => {
	return new Promise<Invitation>((resolve, reject) => {
		makeGRPCCall<CancelInviteRequest, InvitationServiceClient, Invitation>(
			data,
			invitationServiceClient,
			"deleteInvitation"
		)
			.then((response: Invitation) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const addRoleToUser = (
	data: RoleToUserRequest
): Promise<RoleToUserResponse> => {
	return new Promise<RoleToUserResponse>((resolve, reject) => {
		makeGRPCCall<RoleToUserRequest, UserServiceClient, RoleToUserResponse>(
			data,
			userServiceClient,
			"addRoleToUser"
		)
			.then((response: RoleToUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const addUsersToRole = (
	data: UsersToRoleRequest
): Promise<UsersToRoleResponse> => {
	return new Promise<UsersToRoleResponse>((resolve, reject) => {
		makeGRPCCall<UsersToRoleRequest, UserServiceClient, UsersToRoleResponse>(
			data,
			userServiceClient,
			"addUserToRole"
		)
			.then((response: UsersToRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const searchUsers = (
	data: SearchUsersRequest
): Promise<SearchUsersResponse> => {
	return new Promise<SearchUsersResponse>((resolve, reject) => {
		makeGRPCCall<SearchUsersRequest, UserServiceClient, SearchUsersResponse>(
			data,
			userServiceClient,
			"searchUsers"
		)
			.then((response: SearchUsersResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
//WIP: Confirm Remove Role To User
export const removeRoleFromUser = (
	data: RoleToUserRequest
): Promise<RoleToUserResponse> => {
	return new Promise<RoleToUserResponse>((resolve, reject) => {
		makeGRPCCall<RoleToUserRequest, UserServiceClient, RoleToUserResponse>(
			data,
			userServiceClient,
			"removeRoleToUser"
		)
			.then((response: RoleToUserResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readUsers = (
	data: ReadUsersRequest
): Promise<ReadUsersResponse> => {
	return new Promise<ReadUsersResponse>((resolve, reject) => {
		makeGRPCCall<ReadUsersRequest, UserServiceClient, ReadUsersResponse>(
			data,
			userServiceClient,
			"readUsers"
		)
			.then((response: ReadUsersResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createTeam = (
	data: CreateTeamRequest
): Promise<CreateTeamResponse> => {
	return new Promise<CreateTeamResponse>((resolve, reject) => {
		makeGRPCCall<CreateTeamRequest, TeamServiceClient, CreateTeamResponse>(
			data,
			teamServiceClient,
			"createTeam"
		)
			.then((response: CreateTeamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readTeam = (data: ReadTeamRequest): Promise<ReadTeamResponse> => {
	return new Promise<ReadTeamResponse>((resolve, reject) => {
		makeGRPCCall<ReadTeamRequest, TeamServiceClient, ReadTeamResponse>(
			data,
			teamServiceClient,
			"readTeam"
		)
			.then((response: ReadTeamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readTeamList = (
	data: ReadTeamListRequest
): Promise<ReadTeamListResponse> => {
	return new Promise<ReadTeamListResponse>((resolve, reject) => {
		makeGRPCCall<ReadTeamListRequest, TeamServiceClient, ReadTeamListResponse>(
			data,
			teamServiceClient,
			"readTeamsByorganisationId"
		)
			.then((response: ReadTeamListResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const updateTeam = (
	data: UpdateTeamRequest
): Promise<UpdateTeamResponse> => {
	return new Promise<UpdateTeamResponse>((resolve, reject) => {
		makeGRPCCall<UpdateTeamRequest, TeamServiceClient, UpdateTeamResponse>(
			data,
			teamServiceClient,
			"updateTeam"
		)
			.then((response: UpdateTeamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const deleteTeam = (
	data: DeleteTeamRequest
): Promise<DeleteTeamResponse> => {
	return new Promise<DeleteTeamResponse>((resolve, reject) => {
		makeGRPCCall<DeleteTeamRequest, TeamServiceClient, DeleteTeamResponse>(
			data,
			teamServiceClient,
			"deleteTeam"
		)
			.then((response: DeleteTeamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createDepartment = (
	data: CreateDepartmentRequest
): Promise<CreateDepartmentResponse> => {
	return new Promise<CreateDepartmentResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateDepartmentRequest,
			DepartmentServiceClient,
			CreateDepartmentResponse
		>(data, departmentServiceClient, "createDepartment")
			.then((response: CreateDepartmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readDepartment = (
	data: ReadDepartmentRequest
): Promise<ReadDepartmentResponse> => {
	return new Promise<ReadDepartmentResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadDepartmentRequest,
			DepartmentServiceClient,
			ReadDepartmentResponse
		>(data, departmentServiceClient, "readDepartment")
			.then((response: ReadDepartmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readDepartments = (
	data: ReadDepartmentsRequest
): Promise<ReadDepartmentsResponse> => {
	return new Promise<ReadDepartmentsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadDepartmentsRequest,
			DepartmentServiceClient,
			ReadDepartmentsResponse
		>(data, departmentServiceClient, "readDepartments")
			.then((response: ReadDepartmentsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const updateDepartment = (
	data: UpdateDepartmentRequest
): Promise<UpdateDepartmentResponse> => {
	return new Promise<UpdateDepartmentResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateDepartmentRequest,
			DepartmentServiceClient,
			UpdateDepartmentResponse
		>(data, departmentServiceClient, "updateDepartment")
			.then((response: UpdateDepartmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const deleteDepartment = (
	data: DeleteDepartmentRequest
): Promise<DeleteDepartmentResponse> => {
	return new Promise<DeleteDepartmentResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteDepartmentRequest,
			DepartmentServiceClient,
			DeleteDepartmentResponse
		>(data, departmentServiceClient, "deleteDepartment")
			.then((response: DeleteDepartmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readBranchDepartments = (
	data: ReadBranchDepartmentsRequest
): Promise<ReadBranchDepartmentsResponse> => {
	return new Promise<ReadBranchDepartmentsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadBranchDepartmentsRequest,
			DepartmentServiceClient,
			ReadBranchDepartmentsResponse
		>(data, departmentServiceClient, "readBranchDepartments")
			.then((response: ReadBranchDepartmentsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createOrgBranch = (
	data: CreateBranchRequest
): Promise<CreateBranchResponse> => {
	return new Promise<CreateBranchResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateBranchRequest,
			BranchServiceClient,
			CreateBranchResponse
		>(data, branchServiceClient, "createBranch")
			.then((response: CreateBranchResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readBranch = (
	data: ReadBranchRequest
): Promise<ReadBranchResponse> => {
	return new Promise<ReadBranchResponse>((resolve, reject) => {
		makeGRPCCall<ReadBranchRequest, BranchServiceClient, ReadBranchResponse>(
			data,
			branchServiceClient,
			"readBranch"
		)
			.then((response: ReadBranchResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readBranches = (
	data: ReadBranchesRequest
): Promise<ReadBranchesResponse> => {
	return new Promise<ReadBranchesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadBranchesRequest,
			BranchServiceClient,
			ReadBranchesResponse
		>(data, branchServiceClient, "readBranches")
			.then((response: ReadBranchesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateBranch = (
	data: UpdateBranchRequest
): Promise<UpdateBranchResponse> => {
	return new Promise<UpdateBranchResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateBranchRequest,
			BranchServiceClient,
			UpdateBranchResponse
		>(data, branchServiceClient, "updateBranch")
			.then((response: UpdateBranchResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const deleteBranch = (
	data: DeleteBranchRequest
): Promise<DeleteBranchResponse> => {
	return new Promise<DeleteBranchResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteBranchRequest,
			BranchServiceClient,
			DeleteBranchResponse
		>(data, branchServiceClient, "deleteBranch")
			.then((response: DeleteBranchResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createRole = (
	data: CreateRoleRequest
): Promise<CreateRoleResponse> => {
	return new Promise<CreateRoleResponse>((resolve, reject) => {
		makeGRPCCall<CreateRoleRequest, RoleServiceClient, CreateRoleResponse>(
			data,
			roleServiceClient,
			"createRole"
		)
			.then((response: CreateRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readRole = (data: ReadRoleRequest): Promise<ReadRoleResponse> => {
	return new Promise<ReadRoleResponse>((resolve, reject) => {
		makeGRPCCall<ReadRoleRequest, RoleServiceClient, ReadRoleResponse>(
			data,
			roleServiceClient,
			"readRole"
		)
			.then((response: ReadRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readRoles = (
	data: ReadRolesRequest
): Promise<ReadRolesResponse> => {
	return new Promise<ReadRolesResponse>((resolve, reject) => {
		makeGRPCCall<ReadRolesRequest, RoleServiceClient, ReadRolesResponse>(
			data,
			roleServiceClient,
			"readRoles"
		)
			.then((response: ReadRolesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const updateRole = (
	data: UpdateRoleRequest
): Promise<UpdateRoleResponse> => {
	return new Promise<UpdateRoleResponse>((resolve, reject) => {
		makeGRPCCall<UpdateRoleRequest, RoleServiceClient, UpdateRoleResponse>(
			data,
			roleServiceClient,
			"updateRole"
		)
			.then((response: UpdateRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const deleteRole = (
	data: DeleteRoleRequest
): Promise<DeleteRoleResponse> => {
	return new Promise<DeleteRoleResponse>((resolve, reject) => {
		makeGRPCCall<DeleteRoleRequest, RoleServiceClient, DeleteRoleResponse>(
			data,
			roleServiceClient,
			"deleteRole"
		)
			.then((response: DeleteRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readUserRoles = (
	data: ReadRolesRequest
): Promise<ReadRolesResponse> => {
	return new Promise<ReadRolesResponse>((resolve, reject) => {
		makeGRPCCall<ReadRolesRequest, RoleServiceClient, ReadRolesResponse>(
			data,
			roleServiceClient,
			"readUserRoles"
		)
			.then((response: ReadRolesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const addPermissionToRole = (
	data: PermissionToRoleRequest
): Promise<PermissionToRoleResponse> => {
	return new Promise<PermissionToRoleResponse>((resolve, reject) => {
		makeGRPCCall<
			PermissionToRoleRequest,
			RoleServiceClient,
			PermissionToRoleResponse
		>(data, roleServiceClient, "addPermissionToRole")
			.then((response: PermissionToRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const removePermissionFromRole = (
	data: PermissionToRoleRequest
): Promise<PermissionToRoleResponse> => {
	return new Promise<PermissionToRoleResponse>((resolve, reject) => {
		makeGRPCCall<
			PermissionToRoleRequest,
			RoleServiceClient,
			PermissionToRoleResponse
		>(data, roleServiceClient, "removePermissionFromRole")
			.then((response: PermissionToRoleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readPermissions = (
	data: ReadPermissionsRequest
): Promise<ReadPermissionsResponse> => {
	return new Promise<ReadPermissionsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadPermissionsRequest,
			PermissionServiceClient,
			ReadPermissionsResponse
		>(data, permissionServiceClient, "readPermissions")
			.then((response: ReadPermissionsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createShift = (
	data: CreateShiftRequest
): Promise<CreateShiftResponse> => {
	return new Promise<CreateShiftResponse>((resolve, reject) => {
		makeGRPCCall<CreateShiftRequest, ShiftServiceClient, CreateShiftResponse>(
			data,
			shiftServiceClient,
			"createShift"
		)
			.then((response: CreateShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readShift = (
	data: ReadShiftRequest
): Promise<ReadShiftResponse> => {
	return new Promise<ReadShiftResponse>((resolve, reject) => {
		makeGRPCCall<ReadShiftRequest, ShiftServiceClient, ReadShiftResponse>(
			data,
			shiftServiceClient,
			"readShift"
		)
			.then((response: ReadShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const ListShifts = (
	data: ListShiftsRequest
): Promise<ListShiftsResponse> => {
	return new Promise<ListShiftsResponse>((resolve, reject) => {
		makeGRPCCall<ListShiftsRequest, ShiftServiceClient, ListShiftsResponse>(
			data,
			shiftServiceClient,
			"listShifts"
		)
			.then((response: ListShiftsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateShift = (
	data: UpdateShiftRequest
): Promise<UpdateShiftResponse> => {
	return new Promise<UpdateShiftResponse>((resolve, reject) => {
		makeGRPCCall<UpdateShiftRequest, ShiftServiceClient, UpdateShiftResponse>(
			data,
			shiftServiceClient,
			"updateShift"
		)
			.then((response: ReadShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteShift = (
	data: DeleteShiftRequest
): Promise<DeleteShiftResponse> => {
	return new Promise<DeleteShiftResponse>((resolve, reject) => {
		makeGRPCCall<DeleteShiftRequest, ShiftServiceClient, DeleteShiftResponse>(
			data,
			shiftServiceClient,
			"deleteShift"
		)
			.then((response: ReadShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readUserShifts = (
	data: GetUserShiftsRequest
): Promise<GetUserShiftsResponse> => {
	return new Promise<GetUserShiftsResponse>((resolve, reject) => {
		makeGRPCCall<
			GetUserShiftsRequest,
			ShiftServiceClient,
			GetUserShiftsResponse
		>(data, shiftServiceClient, "getUserShifts")
			.then((response: GetUserShiftsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// AssignUsersToShift

export const assignUsersToShift = (
	data: AssignUsersToShiftRequest
): Promise<AssignUsersToShiftResponse> => {
	return new Promise<AssignUsersToShiftResponse>((resolve, reject) => {
		makeGRPCCall<
			AssignUsersToShiftRequest,
			ShiftServiceClient,
			AssignUsersToShiftResponse
		>(data, shiftServiceClient, "assignUsersToShift")
			.then((response: AssignUsersToShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// RemoveUsersFromShift
export const removeUsersFromShift = (
	data: RemoveUsersFromShiftRequest
): Promise<RemoveUsersFromShiftResponse> => {
	return new Promise<RemoveUsersFromShiftResponse>((resolve, reject) => {
		makeGRPCCall<
			RemoveUsersFromShiftRequest,
			ShiftServiceClient,
			RemoveUsersFromShiftResponse
		>(data, shiftServiceClient, "removeUsersFromShift")
			.then((response: RemoveUsersFromShiftResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// ListShiftUsers
export const listShiftUsers = (
	data: ListShiftUsersRequest
): Promise<ListShiftUsersResponse> => {
	return new Promise<ListShiftUsersResponse>((resolve, reject) => {
		makeGRPCCall<
			ListShiftUsersRequest,
			ShiftServiceClient,
			ListShiftUsersResponse
		>(data, shiftServiceClient, "listShiftUsers")
			.then((response: ListShiftUsersResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// Schedule methods
export const readSchedule = (
	data: ReadScheduleRequest
): Promise<ReadScheduleResponse> => {
	return new Promise<ReadScheduleResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadScheduleRequest,
			ShiftScheduleServiceClient,
			ReadScheduleResponse
		>(data, shiftScheduleServiceClient, "readSchedule")
			.then((response: ReadScheduleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateSchedule = (
	data: UpdateScheduleRequest
): Promise<UpdateScheduleResponse> => {
	return new Promise<UpdateScheduleResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateScheduleRequest,
			ShiftScheduleServiceClient,
			UpdateScheduleResponse
		>(data, shiftScheduleServiceClient, "updateSchedule")
			.then((response: UpdateScheduleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteSchedule = (
	data: DeleteScheduleRequest
): Promise<DeleteScheduleResponse> => {
	return new Promise<DeleteScheduleResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteScheduleRequest,
			ShiftScheduleServiceClient,
			DeleteScheduleResponse
		>(data, shiftScheduleServiceClient, "deleteSchedule")
			.then((response: DeleteScheduleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const listSchedules = (
	data: ListSchedulesRequest
): Promise<ListSchedulesResponse> => {
	return new Promise<ListSchedulesResponse>((resolve, reject) => {
		makeGRPCCall<
			ListSchedulesRequest,
			ShiftScheduleServiceClient,
			ListSchedulesResponse
		>(data, shiftScheduleServiceClient, "listSchedules")
			.then((response: ListSchedulesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// Clock methods
export const clockIn = (data: ClockInRequest): Promise<ClockInResponse> => {
	return new Promise<ClockInResponse>((resolve, reject) => {
		makeGRPCCall<ClockInRequest, ShiftScheduleServiceClient, ClockInResponse>(
			data,
			shiftScheduleServiceClient,
			"clockIn"
		)
			.then((response: ClockInResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const clockOut = (data: ClockOutRequest): Promise<ClockOutResponse> => {
	return new Promise<ClockOutResponse>((resolve, reject) => {
		makeGRPCCall<ClockOutRequest, ShiftScheduleServiceClient, ClockOutResponse>(
			data,
			shiftScheduleServiceClient,
			"clockOut"
		)
			.then((response: ClockOutResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// Break methods
export const createBreak = (
	data: CreateBreakRequest
): Promise<CreateBreakResponse> => {
	return new Promise<CreateBreakResponse>((resolve, reject) => {
		makeGRPCCall<CreateBreakRequest, BreakServiceClient, CreateBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"createBreak"
		)
			.then((response: CreateBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readBreak = (
	data: ReadBreakRequest
): Promise<ReadBreakResponse> => {
	return new Promise<ReadBreakResponse>((resolve, reject) => {
		makeGRPCCall<ReadBreakRequest, BreakServiceClient, ReadBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"readBreak"
		)
			.then((response: ReadBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateBreak = (
	data: UpdateBreakRequest
): Promise<UpdateBreakResponse> => {
	return new Promise<UpdateBreakResponse>((resolve, reject) => {
		makeGRPCCall<UpdateBreakRequest, BreakServiceClient, UpdateBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"updateBreak"
		)
			.then((response: UpdateBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteBreak = (
	data: DeleteBreakRequest
): Promise<DeleteBreakResponse> => {
	return new Promise<DeleteBreakResponse>((resolve, reject) => {
		makeGRPCCall<DeleteBreakRequest, BreakServiceClient, DeleteBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"deleteBreak"
		)
			.then((response: DeleteBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const listBreaks = (
	data: ListBreaksRequest
): Promise<ListBreaksResponse> => {
	return new Promise<ListBreaksResponse>((resolve, reject) => {
		makeGRPCCall<ListBreaksRequest, BreakServiceClient, ListBreaksResponse>(
			data,
			scheduleBreakServiceClient,
			"listBreaks"
		)
			.then((response: ListBreaksResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const takeBreak = (
	data: TakeBreakRequest
): Promise<TakeBreakResponse> => {
	return new Promise<TakeBreakResponse>((resolve, reject) => {
		makeGRPCCall<TakeBreakRequest, BreakServiceClient, TakeBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"takeBreak"
		)
			.then((response: TakeBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const resumeBreak = (
	data: ResumeBreakRequest
): Promise<ResumeBreakResponse> => {
	return new Promise<ResumeBreakResponse>((resolve, reject) => {
		makeGRPCCall<ResumeBreakRequest, BreakServiceClient, ResumeBreakResponse>(
			data,
			scheduleBreakServiceClient,
			"resumeBreak"
		)
			.then((response: ResumeBreakResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readContacts = (
	data: ReadContactsRequest
): Promise<ReadContactsResponse> => {
	return new Promise<ReadContactsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadContactsRequest,
			ContactServiceClient,
			ReadContactsResponse
		>(data, contactServiceClient, "readContacts")
			.then((response: ReadContactsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readContact = (
	data: ReadContactRequest
): Promise<ReadContactResponse> => {
	return new Promise<ReadContactResponse>((resolve, reject) => {
		makeGRPCCall<ReadContactRequest, ContactServiceClient, ReadContactResponse>(
			data,
			contactServiceClient,
			"readContact"
		)
			.then((response: ReadContactResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createContact = (
	data: CreateContactRequest
): Promise<CreateContactResponse> => {
	return new Promise<CreateContactResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateContactRequest,
			ContactServiceClient,
			CreateContactResponse
		>(data, contactServiceClient, "createContact")
			.then((response: CreateContactResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateContact = (
	data: UpdateContactRequest
): Promise<UpdateContactResponse> => {
	return new Promise<UpdateContactResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateContactRequest,
			ContactServiceClient,
			UpdateContactResponse
		>(data, contactServiceClient, "updateContact")
			.then((response: UpdateContactResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteContact = (
	data: DeleteContactRequest
): Promise<DeleteContactResponse> => {
	return new Promise<DeleteContactResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteContactRequest,
			ContactServiceClient,
			DeleteContactResponse
		>(data, contactServiceClient, "deleteContact")
			.then((response: DeleteContactResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const bulkContactsImport = (data: BulkContactsImportRequest) => {
	return new Promise<BulkContactsImportResponse>((resolve, reject) => {
		makeGRPCCall<
			BulkContactsImportRequest,
			ContactServiceClient,
			BulkContactsImportResponse
		>(data, contactServiceClient, "bulkContactsImport")
			.then((response: BulkContactsImportResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const bulkContactsExport = (
	data: BulkContactsExportRequest
): Promise<Blob> => {
	return new Promise<Blob>((resolve, reject) => {
		const call = contactServiceClient.getService().bulkContactsExport(data, {
			meta: {
				authorization: tokenObject().authorization,
			},
		});

		const chunks: Uint8Array[] = [];

		// Handle stream events
		call.responses.onMessage((response: BulkContactsExportResponse) => {
			chunks.push(response.chunkData);
		});

		call.responses.onComplete(() => {
			resolve(
				new Blob(
					chunks.map((chunk) => new Uint8Array(chunk)),
					{ type: "text/csv" }
				)
			);
		});

		call.responses.onError((reason: Error) => {
			reject(reason);
		});
	});
};

export const readCustomFieldDefinitions = (
	data: ReadCustomFieldDefinitionsRequest
): Promise<ReadCustomFieldDefinitionsResponse> => {
	return new Promise<ReadCustomFieldDefinitionsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadCustomFieldDefinitionsRequest,
			ContactServiceClient,
			ReadCustomFieldDefinitionsResponse
		>(data, contactServiceClient, "readCustomFieldDefinitions")
			.then((response: ReadCustomFieldDefinitionsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const findContactDuplicates = (
	data: FindContactDuplicatesRequest
): Promise<FindContactDuplicatesResponse> => {
	return new Promise<FindContactDuplicatesResponse>((resolve, reject) => {
		makeGRPCCall<
			FindContactDuplicatesRequest,
			ContactServiceClient,
			FindContactDuplicatesResponse
		>(data, contactServiceClient, "findContactDuplicates")
			.then((response: FindContactDuplicatesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const mergeContacts = (data: MergeDuplicateContactsRequest) => {
	return new Promise<MergeDuplicateContactsResponse>((resolve, reject) => {
		makeGRPCCall<
			MergeDuplicateContactsRequest,
			ContactServiceClient,
			MergeDuplicateContactsResponse
		>(data, contactServiceClient, "mergeDuplicateContacts")
			.then((response: MergeDuplicateContactsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readGroup = (
	data: ReadGroupRequest
): Promise<ReadGroupResponse> => {
	return new Promise<ReadGroupResponse>((resolve, reject) => {
		makeGRPCCall<ReadGroupRequest, GroupServiceClient, ReadGroupResponse>(
			data,
			groupServiceClient,
			"readGroup"
		)
			.then((response: ReadGroupResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readGroups = (
	data: ReadGroupsRequest
): Promise<ReadGroupsResponse> => {
	return new Promise<ReadGroupsResponse>((resolve, reject) => {
		makeGRPCCall<ReadGroupsRequest, GroupServiceClient, ReadGroupsResponse>(
			data,
			groupServiceClient,
			"readGroups"
		)
			.then((response: ReadGroupsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createGroup = (
	data: CreateGroupRequest
): Promise<CreateGroupResponse> => {
	return new Promise<CreateGroupResponse>((resolve, reject) => {
		makeGRPCCall<CreateGroupRequest, GroupServiceClient, CreateGroupResponse>(
			data,
			groupServiceClient,
			"createGroup"
		)
			.then((response: CreateGroupResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateGroup = (
	data: UpdateGroupRequest
): Promise<UpdateGroupResponse> => {
	return new Promise<UpdateGroupResponse>((resolve, reject) => {
		makeGRPCCall<UpdateGroupRequest, GroupServiceClient, UpdateGroupResponse>(
			data,
			groupServiceClient,
			"updateGroup"
		)
			.then((response: UpdateGroupResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteGroup = (
	data: DeleteGroupRequest
): Promise<DeleteGroupResponse> => {
	return new Promise<DeleteGroupResponse>((resolve, reject) => {
		makeGRPCCall<DeleteGroupRequest, GroupServiceClient, DeleteGroupResponse>(
			data,
			groupServiceClient,
			"deleteGroup"
		)
			.then((response: DeleteGroupResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readGroupContacts = (
	data: ReadGroupContactsRequest
): Promise<ReadGroupContactsResponse> => {
	return new Promise<ReadGroupContactsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadGroupContactsRequest,
			GroupServiceClient,
			ReadGroupContactsResponse
		>(data, groupServiceClient, "readGroupContacts")
			.then((response: ReadGroupContactsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createApp = (
	data: CreateAppRequest
): Promise<CreateAppResponse> => {
	return new Promise<CreateAppResponse>((resolve, reject) => {
		makeGRPCCall<CreateAppRequest, AppServiceClient, CreateAppResponse>(
			data,
			appServiceClient,
			"createApp"
		)
			.then((response: CreateAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const updateApp = (
	data: UpdateAppRequest
): Promise<UpdateAppResponse> => {
	return new Promise<UpdateAppResponse>((resolve, reject) => {
		makeGRPCCall<UpdateAppRequest, AppServiceClient, UpdateAppResponse>(
			data,
			appServiceClient,
			"updateApp"
		)
			.then((response: UpdateAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const deleteApp = (
	data: DeleteAppRequest
): Promise<DeleteAppResponse> => {
	return new Promise<DeleteAppResponse>((resolve, reject) => {
		makeGRPCCall<DeleteAppRequest, AppServiceClient, DeleteAppResponse>(
			data,
			appServiceClient,
			"deleteApp"
		)
			.then((response: DeleteAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readApp = (data: ReadAppRequest): Promise<ReadAppResponse> => {
	return new Promise<ReadAppResponse>((resolve, reject) => {
		makeGRPCCall<ReadAppRequest, AppServiceClient, ReadAppResponse>(
			data,
			appServiceClient,
			"readApp"
		)
			.then((response: ReadAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readApps = (data: ReadAppsRequest): Promise<ReadAppsResponse> => {
	return new Promise<ReadAppsResponse>((resolve, reject) => {
		makeGRPCCall<ReadAppsRequest, AppServiceClient, ReadAppsResponse>(
			data,
			appServiceClient,
			"readApps"
		)
			.then((response: ReadAppsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const approveApp = (
	data: ApproveAppRequest
): Promise<ApproveAppResponse> => {
	return new Promise<ApproveAppResponse>((resolve, reject) => {
		makeGRPCCall<ApproveAppRequest, AppServiceClient, ApproveAppResponse>(
			data,
			appServiceClient,
			"approveApp"
		)
			.then((response: ApproveAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createTemplateApp = (
	data: CreateTemplateAppRequest
): Promise<CreateTemplateAppResponse> => {
	return new Promise<CreateTemplateAppResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTemplateAppRequest,
			AppServiceClient,
			CreateTemplateAppResponse
		>(data, appServiceClient, "createTemplateApp")
			.then((response: CreateTemplateAppResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const createAppWithTemplate = (
	data: CreateAppWithTemplateRequest
): Promise<CreateAppWithTemplateResponse> => {
	return new Promise<CreateAppWithTemplateResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateAppWithTemplateRequest,
			AppServiceClient,
			CreateAppWithTemplateResponse
		>(data, appServiceClient, "createAppWithTemplate")
			.then((response: CreateAppWithTemplateResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
export const readAppTemplates = (
	data: ReadAppTemplatesRequest
): Promise<ReadAppTemplatesResponse> => {
	return new Promise<ReadAppTemplatesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadAppTemplatesRequest,
			AppServiceClient,
			ReadAppTemplatesResponse
		>(data, appServiceClient, "readAppTemplates")
			.then((response: ReadAppTemplatesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readAuthContext = (
	data: ReadAuthContextRequest
): Promise<ReadAuthContextResponse> => {
	return new Promise<ReadAuthContextResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadAuthContextRequest,
			AuthServiceClient,
			ReadAuthContextResponse
		>(data, authServiceClient, "readAuthContext")
			.then((response: ReadAuthContextResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createLeaveType = (
	data: CreateLeaveTypeRequest
): Promise<CreateLeaveTypeResponse> => {
	return new Promise<CreateLeaveTypeResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateLeaveTypeRequest,
			LeaveServiceClient,
			CreateLeaveTypeResponse
		>(data, leaveServiceClient, "createLeaveType")
			.then((response: CreateLeaveTypeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateLeaveType = (
	data: UpdateLeaveTypeRequest
): Promise<UpdateLeaveTypeResponse> => {
	return new Promise<UpdateLeaveTypeResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateLeaveTypeRequest,
			LeaveServiceClient,
			UpdateLeaveTypeResponse
		>(data, leaveServiceClient, "updateLeaveType")
			.then((response: UpdateLeaveTypeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveType = (
	data: ReadLeaveTypeRequest
): Promise<ReadLeaveTypeResponse> => {
	return new Promise<ReadLeaveTypeResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveTypeRequest,
			LeaveServiceClient,
			ReadLeaveTypeResponse
		>(data, leaveServiceClient, "readLeaveType")
			.then((response: ReadLeaveTypeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveTypes = (
	data: ReadLeaveTypesRequest
): Promise<ReadLeaveTypesResponse> => {
	return new Promise<ReadLeaveTypesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveTypesRequest,
			LeaveServiceClient,
			ReadLeaveTypesResponse
		>(data, leaveServiceClient, "readLeaveTypes")
			.then((response: ReadLeaveTypesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteLeaveType = (
	data: DeleteLeaveTypeRequest
): Promise<DeleteLeaveTypeResponse> => {
	return new Promise<DeleteLeaveTypeResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteLeaveTypeRequest,
			LeaveServiceClient,
			DeleteLeaveTypeResponse
		>(data, leaveServiceClient, "deleteLeaveType")
			.then((response: DeleteLeaveTypeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createLeaveRequest = (
	data: CreateLeaveRequestRequest
): Promise<CreateLeaveRequestResponse> => {
	return new Promise<CreateLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateLeaveRequestRequest,
			LeaveServiceClient,
			CreateLeaveRequestResponse
		>(data, leaveServiceClient, "createLeaveRequest")
			.then((response: CreateLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveRequest = (
	data: ReadLeaveRequestRequest
): Promise<ReadLeaveRequestResponse> => {
	return new Promise<ReadLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveRequestRequest,
			LeaveServiceClient,
			ReadLeaveRequestResponse
		>(data, leaveServiceClient, "readLeaveRequest")
			.then((response: ReadLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveRequests = (
	data: ReadLeaveRequestsRequest
): Promise<ReadLeaveRequestsResponse> => {
	return new Promise<ReadLeaveRequestsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveRequestsRequest,
			LeaveServiceClient,
			ReadLeaveRequestsResponse
		>(data, leaveServiceClient, "readLeaveRequests")
			.then((response: ReadLeaveRequestsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readMyLeaveRequests = (
	data: ReadMyLeaveRequestsRequest
): Promise<ReadMyLeaveRequestsResponse> => {
	return new Promise<ReadMyLeaveRequestsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadMyLeaveRequestsRequest,
			LeaveServiceClient,
			ReadMyLeaveRequestsResponse
		>(data, leaveServiceClient, "readMyLeaveRequests")
			.then((response: ReadMyLeaveRequestsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const approveLeaveRequest = (
	data: ApproveLeaveRequestRequest
): Promise<ApproveLeaveRequestResponse> => {
	return new Promise<ApproveLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			ApproveLeaveRequestRequest,
			LeaveServiceClient,
			ApproveLeaveRequestResponse
		>(data, leaveServiceClient, "approveLeaveRequest")
			.then((response: ApproveLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const rejectLeaveRequest = (
	data: RejectLeaveRequestRequest
): Promise<RejectLeaveRequestResponse> => {
	return new Promise<RejectLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			RejectLeaveRequestRequest,
			LeaveServiceClient,
			RejectLeaveRequestResponse
		>(data, leaveServiceClient, "rejectLeaveRequest")
			.then((response: RejectLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const cancelLeaveRequest = (
	data: CancelLeaveRequestRequest
): Promise<CancelLeaveRequestResponse> => {
	return new Promise<CancelLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			CancelLeaveRequestRequest,
			LeaveServiceClient,
			CancelLeaveRequestResponse
		>(data, leaveServiceClient, "cancelLeaveRequest")
			.then((response: CancelLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const endLeaveRequest = (
	data: EndLeaveRequestRequest
): Promise<EndLeaveRequestResponse> => {
	return new Promise<EndLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			EndLeaveRequestRequest,
			LeaveServiceClient,
			EndLeaveRequestResponse
		>(data, leaveServiceClient, "endLeaveRequest")
			.then((response: EndLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const appealLeaveRequest = (
	data: AppealLeaveRequestRequest
): Promise<AppealLeaveRequestResponse> => {
	return new Promise<AppealLeaveRequestResponse>((resolve, reject) => {
		makeGRPCCall<
			AppealLeaveRequestRequest,
			LeaveServiceClient,
			AppealLeaveRequestResponse
		>(data, leaveServiceClient, "appealLeaveRequest")
			.then((response: AppealLeaveRequestResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveBalance = (
	data: ReadLeaveBalanceRequest
): Promise<ReadLeaveBalanceResponse> => {
	return new Promise<ReadLeaveBalanceResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveBalanceRequest,
			LeaveServiceClient,
			ReadLeaveBalanceResponse
		>(data, leaveServiceClient, "readLeaveBalance")
			.then((response: ReadLeaveBalanceResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readLeaveRequestsMetrics = (
	data: ReadLeaveRequestsMetricsRequest
): Promise<ReadLeaveRequestsMetricsResponse> => {
	return new Promise<ReadLeaveRequestsMetricsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadLeaveRequestsMetricsRequest,
			LeaveServiceClient,
			ReadLeaveRequestsMetricsResponse
		>(data, leaveServiceClient, "readLeaveRequestsMetrics")
			.then((response: ReadLeaveRequestsMetricsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
