import { RpcError } from "grpc-web";
import { makeGRPCCall } from "./rpcUtils";


import {
	businessRuleClient
} from "../globalClient/GlobalClients";
import type { ReadBusinessRulesRequest, ReadBusinessRulesResponse, ReadBusinessRuleRequest, ReadBusinessRuleResponse, CreateBusinessRuleRequest, CreateBusinessRuleResponse, UpdateBusinessRuleRequest, UpdateBusinessRuleResponse, DeleteBusinessRuleRequest, DeleteBusinessRuleResponse } from "@/pb/business_rule";
import type { BusinessRuleServiceClient } from "@/pb/business_rule.client";


export const readBusinessRules = async (
	data: ReadBusinessRulesRequest
): Promise<ReadBusinessRulesResponse> => {
	return new Promise<ReadBusinessRulesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadBusinessRulesRequest,
			BusinessRuleServiceClient,
			ReadBusinessRulesResponse
		>(data, businessRuleClient, "readBusinessRules")
			.then((response: ReadBusinessRulesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readBusinessRule = async (
	data: ReadBusinessRuleRequest
): Promise<ReadBusinessRuleResponse> => {
	return new Promise<ReadBusinessRuleResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadBusinessRuleRequest,
			BusinessRuleServiceClient,
			ReadBusinessRuleResponse
		>(data, businessRuleClient, "readBusinessRule")
			.then((response: ReadBusinessRuleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createBusinessRule = async (
	data: CreateBusinessRuleRequest
): Promise<CreateBusinessRuleResponse> => {
	return new Promise<CreateBusinessRuleResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateBusinessRuleRequest,
			BusinessRuleServiceClient,
			CreateBusinessRuleResponse
		>(data, businessRuleClient, "createBusinessRule")
			.then((response: CreateBusinessRuleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateBusinessRule = async (
	data: UpdateBusinessRuleRequest
): Promise<UpdateBusinessRuleResponse> => {
	return new Promise<UpdateBusinessRuleResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateBusinessRuleRequest,
			BusinessRuleServiceClient,
			UpdateBusinessRuleResponse
		>(data, businessRuleClient, "updateBusinessRule")
			.then((response: UpdateBusinessRuleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteBusinessRule = async (
	data: DeleteBusinessRuleRequest
): Promise<DeleteBusinessRuleResponse> => {
	return new Promise<DeleteBusinessRuleResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteBusinessRuleRequest,
			BusinessRuleServiceClient,
			DeleteBusinessRuleResponse
		>(data, businessRuleClient, "deleteBusinessRule")
			.then((response: DeleteBusinessRuleResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};