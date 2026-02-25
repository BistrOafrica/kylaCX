import { RpcError } from "grpc-web";
import {
	integrationServiceClient,
	whatsappEmbeddedSignUpServiceClient,
} from "../globalClient/GlobalClients";
import type { SendEmbeddedSignUpEventRequest, SendEmbeddedSignUpEventResponse, CreateFacebookIntegrationRequest, CreateFacebookIntegrationResponse, ReadIntegrationsRequest, ReadIntegrationsResponse, CreateAfricasTalkingSmsIntegrationRequest, CreateAfricasTalkingSmsIntegrationResponse, CreateImapIntegrationRequest, CreateImapIntegrationResponse, UpdateImapIntegrationRequest, UpdateImapIntegrationResponse, TestImapConnectionRequest, TestImapConnectionResponse, TestSmtpConnectionRequest, TestSmtpConnectionResponse, DeleteIntegrationRequest, DeleteIntegrationResponse } from "@/pb/integration";
import type { WhatsAppEmbeddedSignUpServiceClient, IntegrationServiceClient } from "@/pb/integration.client";
import type { ReadSourceRequest, ReadSourceResponse, ReadSourcesRequest, ReadSourcesResponse } from "@/pb/source";
import { makeGRPCCall } from "./rpcUtils";


export const createWhatsappIntegration = (
	data: SendEmbeddedSignUpEventRequest,
): Promise<SendEmbeddedSignUpEventResponse> => {
	return new Promise<SendEmbeddedSignUpEventResponse>((resolve, reject) => {
		makeGRPCCall<
			SendEmbeddedSignUpEventRequest,
			WhatsAppEmbeddedSignUpServiceClient,
			SendEmbeddedSignUpEventResponse
		>(data, whatsappEmbeddedSignUpServiceClient, "sendSignUpEvent")
			.then((response: SendEmbeddedSignUpEventResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createFacebookIntegration = (
	data: CreateFacebookIntegrationRequest,
): Promise<CreateFacebookIntegrationResponse> => {
	return new Promise<CreateFacebookIntegrationResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateFacebookIntegrationRequest,
			IntegrationServiceClient,
			CreateFacebookIntegrationResponse
		>(data, integrationServiceClient, "createFacebookIntegration")
			.then((response: CreateFacebookIntegrationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readIntegrations = (
	data: ReadIntegrationsRequest,
): Promise<ReadIntegrationsResponse> => {
	return new Promise<ReadIntegrationsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadIntegrationsRequest,
			IntegrationServiceClient,
			ReadIntegrationsResponse
		>(data, integrationServiceClient, "readIntegrations")
			.then((response: ReadIntegrationsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createAfricasTalkingSmsIntegration = (
	data: CreateAfricasTalkingSmsIntegrationRequest,
): Promise<CreateAfricasTalkingSmsIntegrationResponse> => {
	return new Promise<CreateFacebookIntegrationResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateAfricasTalkingSmsIntegrationRequest,
			IntegrationServiceClient,
			CreateFacebookIntegrationResponse
		>(data, integrationServiceClient, "createAfricasTalkingSmsIntegration")
			.then((response: CreateFacebookIntegrationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readSource = (
	data: ReadSourceRequest,
): Promise<ReadSourceResponse> => {
	return new Promise<ReadSourceResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadSourceRequest,
			IntegrationServiceClient,
			ReadSourceResponse
		>(data, integrationServiceClient, "readSource")
			.then((response: ReadSourceResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readSources = (
	data: ReadSourcesRequest,
): Promise<ReadSourcesResponse> => {
	return new Promise<ReadSourcesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadSourcesRequest,
			IntegrationServiceClient,
			ReadSourcesResponse
		>(data, integrationServiceClient, "readSources")
			.then((response: ReadSourcesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createImapIntegration = (
	data: CreateImapIntegrationRequest,
): Promise<CreateImapIntegrationResponse> => {
	return new Promise<CreateImapIntegrationResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateImapIntegrationRequest,
			IntegrationServiceClient,
			CreateImapIntegrationResponse
		>(data, integrationServiceClient, "createImapIntegration")
			.then((response: CreateImapIntegrationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateImapIntegration = (
	data: UpdateImapIntegrationRequest,
): Promise<UpdateImapIntegrationResponse> => {
	return new Promise<UpdateImapIntegrationResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateImapIntegrationRequest,
			IntegrationServiceClient,
			UpdateImapIntegrationResponse
		>(data, integrationServiceClient, "updateImapIntegration")
			.then((response: UpdateImapIntegrationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const testImapConnection = (
	data: TestImapConnectionRequest,
): Promise<TestImapConnectionResponse> => {
	return new Promise<TestImapConnectionResponse>((resolve, reject) => {
		makeGRPCCall<
			TestImapConnectionRequest,
			IntegrationServiceClient,
			TestImapConnectionResponse
		>(data, integrationServiceClient, "testImapConnection")
			.then((response: TestImapConnectionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const testSmtpConnection = (
	data: TestSmtpConnectionRequest,
): Promise<TestSmtpConnectionResponse> => {
	return new Promise<TestSmtpConnectionResponse>((resolve, reject) => {
		makeGRPCCall<
			TestSmtpConnectionRequest,
			IntegrationServiceClient,
			TestSmtpConnectionResponse
		>(data, integrationServiceClient, "testSmtpConnection")
			.then((response: TestSmtpConnectionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteIntegration = (
	data: DeleteIntegrationRequest,
): Promise<DeleteIntegrationResponse> => {
	return new Promise<DeleteIntegrationResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteIntegrationRequest,
			IntegrationServiceClient,
			DeleteIntegrationResponse
		>(data, integrationServiceClient, "deleteIntegration")
			.then((response: DeleteIntegrationResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
