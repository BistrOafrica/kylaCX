import { RpcError } from "grpc-web";
import {
	attachmentServiceClient,
	escalationThreadServiceClient,
	faqServiceClient,
	slaServiceClient,
	threadServiceClient,
	ticketAnalyticsServiceClient,
	ticketAssignmentPolicyServiceClient,
	ticketCategoryServiceClient,
	ticketCustomFieldDefinitionServiceClient,
	ticketDashboardServiceClient,
	ticketLogServiceClient,
	ticketMacroServiceClient,
	ticketNoteServiceClient,
	ticketRoomServiceClient,
	ticketScriptServiceClient,
	ticketServiceClient,
	ticketTagServiceClient,
} from "../globalClient/GlobalClients";
import type { CreateAttachmentRequest, CreateAttachmentResponse, ReadAttachmentRequest, ReadAttachmentResponse, ReadAttachmentsRequest, ReadAttachmentsResponse, GetAttachmentPresignedURLRequest, GetAttachmentPresignedURLResponse } from "@/pb/attachment";
import type { AttachmentServiceClient } from "@/pb/attachment.client";
import type { CreateTicketCategoryRequest, CreateTicketCategoryResponse, UpdateTicketCategoryRequest, UpdateTicketCategoryResponse, ReadTicketCategoryRequest, ReadTicketCategoryResponse, ReadTicketCategoriesRequest, ReadTicketCategoriesResponse, DeleteTicketCategoryRequest, DeleteTicketCategoryResponse } from "@/pb/category";
import type { TicketCategoryServiceClient } from "@/pb/category.client";
import type { CreateEscalationThreadRequest, CreateEscalationThreadResponse } from "@/pb/escalation_thread";
import type { EscalationThreadServiceClient } from "@/pb/escalation_thread.client";
import type { CreateFaqRequest, CreateFaqResponse, UpdateFaqRequest, UpdateFaqResponse, ReadFaqRequest, ReadFaqResponse, ReadFaqsRequest, ReadFaqsResponse, DeleteFaqRequest, DeleteFaqResponse } from "@/pb/faq";
import type { FaqServiceClient } from "@/pb/faq.client";
import type { CreateSLAPolicyRequest, CreateSLAPolicyResponse, UpdateSLAPolicyRequest, UpdateSLAPolicyResponse, DeleteSLAPolicyRequest, DeleteSLAPolicyResponse, ListSLAPoliciesRequest, ListSLAPoliciesResponse, GetTicketSLARequest, GetTicketSLAResponse } from "@/pb/sla";
import type { SLAServiceClient } from "@/pb/sla.client";
import type { CreateThreadRequest, CreateThreadResponse } from "@/pb/thread";
import type { ThreadServiceClient } from "@/pb/thread.client";
import type { ReadTicketsInRoomRequest, ReadTicketsInRoomResponse, ChangeTicketsCategoryRequest, ChangeTicketsCategoryResponse, ChangeTicketsPriorityRequest, ChangeTicketsPriorityResponse, ChangeTicketsStatusRequest, ChangeTicketsStatusResponse, ChangeTicketsTagsRequest, ChangeTicketsTagsResponse, MarkTicketsAsReadRequest, MarkTicketsAsReadResponse, MarkTicketsAsUnreadRequest, MarkTicketsAsUnreadResponse, MarkTicketsAsSpamRequest, MarkTicketsAsSpamResponse, MarkTicketsAsNotSpamRequest, MarkTicketsAsNotSpamResponse, ArchiveTicketsRequest, ArchiveTicketsResponse, UnarchiveTicketsRequest, UnarchiveTicketsResponse, ChangeTicketsAssignedAgentRequest, ChangeTicketsAssignedAgentResponse, ReadTicketsRequest, ReadTicketsResponse, ReadMyAssignedTicketsRequest, ReadMyAssignedTicketsResponse, ReadUnassignedTicketsRequest, ReadUnassignedTicketsResponse, ReadArchivedTicketsRequest, ReadArchivedTicketsResponse, ReadSpamTicketsRequest, ReadSpamTicketsResponse, CreateTicketRequest, CreateTicketResponse, UpdateTicketRequest, UpdateTicketResponse, ChangeTicketsFollowersRequest, ChangeTicketsFollowersResponse, MergeTicketsRequest, MergeTicketsResponse, GetPotentialDuplicatesRequest, GetPotentialDuplicatesResponse, ReviewDuplicateRequest, ReviewDuplicateResponse, GetMergeHistoryRequest, GetMergeHistoryResponse, GetMergedTicketsRequest, GetMergedTicketsResponse, GetMasterTicketRequest, GetMasterTicketResponse, ReadTicketRequest, ReadTicketResponse } from "@/pb/ticket";
import type { TicketServiceClient } from "@/pb/ticket.client";
import type { GetKPIMetricsRequest, KPIMetrics, GetTicketVolumeRequest, TicketVolumeResponse, GetChannelDistributionRequest, ChannelDistributionResponse, GetIntegrationDistributionRequest, IntegrationDistributionResponse, GetStatusDistributionRequest, StatusDistributionResponse, GetPriorityDistributionRequest, PriorityDistributionResponse, GetAgentPerformanceRequest, AgentPerformanceResponse, GetSLAComplianceRequest, SLAComplianceResponse, GetActivityHeatmapRequest, ActivityHeatmapResponse, GetResolutionTimeAnalyticsRequest, ResolutionTimeResponse, GetEscalationMetricsRequest, EscalationMetrics, GetLiveTicketTrackingRequest, LiveTicketTrackingResponse, GetTrendAnalysisRequest, TrendAnalysisResponse, GetSentimentAnalyticsRequest, SentimentAnalyticsResponse } from "@/pb/ticket_analytics";
import type { TicketAnalyticsServiceClient } from "@/pb/ticket_analytics.client";
import type { CreateTicketAssignmentPolicyRequest, CreateTicketAssignmentPolicyResponse, UpdateTicketAssignmentPolicyRequest, UpdateTicketAssignmentPolicyResponse, DeleteTicketAssignmentPolicyRequest, DeleteTicketAssignmentPolicyResponse, ReadTicketAssignmentPolicyRequest, ReadTicketAssignmentPolicyResponse, ReadTicketAssignmentPoliciesRequest, ReadTicketAssignmentPoliciesResponse, ReadOwnerAssignmentPolicyRequest, ReadOwnerAssignmentPolicyResponse } from "@/pb/ticket_assignment_rule";
import type { TicketAssignmentPolicyServiceClient } from "@/pb/ticket_assignment_rule.client";
import type { CreateTicketCustomFieldDefinitionRequest, CreateTicketCustomFieldDefinitionResponse, UpdateTicketCustomFieldDefinitionRequest, UpdateTicketCustomFieldDefinitionResponse, ReadTicketCustomFieldDefinitionsRequest, ReadTicketCustomFieldDefinitionsResponse, DeleteTicketCustomFieldDefinitionRequest, DeleteTicketCustomFieldDefinitionResponse } from "@/pb/ticket_custom_field_definition";
import type { TicketCustomFieldDefinitionServiceClient } from "@/pb/ticket_custom_field_definition.client";
import type { CreateDashboardRequest, DashboardResponse, GetDashboardRequest, UpdateDashboardRequest, DeleteDashboardRequest, DeleteDashboardResponse, ListDashboardsRequest, ListDashboardsResponse, DuplicateDashboardRequest, GetAvailableMetricsRequest, GetAvailableMetricsResponse, ValidateDashboardRequest, ValidateDashboardResponse, ValidateMetricConfigRequest, ValidateMetricConfigResponse } from "@/pb/ticket_dashboards";
import type { TicketDashboardServiceClient } from "@/pb/ticket_dashboards.client";
import type { ReadTicketLogsRequest, ReadTicketLogsResponse } from "@/pb/ticket_log";
import type { TicketLogServiceClient } from "@/pb/ticket_log.client";
import type { CreateTicketMacroRequest, CreateTicketMacroResponse, UpdateTicketMacroRequest, UpdateTicketMacroResponse, ReadTicketMacroRequest, ReadTicketMacroResponse, ReadTicketMacrosRequest, ReadTicketMacrosResponse, DeleteTicketMacroRequest, DeleteTicketMacroResponse } from "@/pb/ticket_macro";
import type { TicketMacroServiceClient } from "@/pb/ticket_macro.client";
import type { CreateTicketNoteRequest, CreateTicketNoteResponse, UpdateTicketNoteRequest, UpdateTicketNoteResponse, ReadTicketNoteRequest, ReadTicketNoteResponse, ReadTicketNotesRequest, ReadTicketNotesResponse, DeleteTicketNoteRequest, DeleteTicketNoteResponse } from "@/pb/ticket_note";
import type { TicketNoteServiceClient } from "@/pb/ticket_note.client";
import type { CreateTicketRoomRequest, CreateTicketRoomResponse, UpdateTicketRoomRequest, UpdateTicketRoomResponse, ReadTicketRoomRequest, ReadTicketRoomResponse, ReadTicketRoomsRequest, ReadTicketRoomsResponse, DeleteTicketRoomRequest, DeleteTicketRoomResponse } from "@/pb/ticket_room";
import type { TicketRoomServiceClient } from "@/pb/ticket_room.client";
import type { CreateTicketScriptRequest, CreateTicketScriptResponse, UpdateTicketScriptRequest, UpdateTicketScriptResponse, ReadTicketScriptRequest, ReadTicketScriptResponse, ReadTicketScriptsRequest, ReadTicketScriptsResponse, DeleteTicketScriptRequest, DeleteTicketScriptResponse } from "@/pb/ticket_script";
import type { TicketScriptServiceClient } from "@/pb/ticket_script.client";
import type { CreateTicketTagRequest, CreateTicketTagResponse, UpdateTicketTagRequest, UpdateTicketTagResponse, ReadTicketTagRequest, ReadTicketTagResponse, ReadTicketTagsRequest, ReadTicketTagsResponse, DeleteTicketTagRequest, DeleteTicketTagResponse } from "@/pb/ticket_tag";
import type { TicketTagServiceClient } from "@/pb/ticket_tag.client";
import { makeGRPCCall } from "./rpcUtils";

export const createFaq = (
	data: CreateFaqRequest
): Promise<CreateFaqResponse> => {
	return new Promise<CreateFaqResponse>((resolve, reject) => {
		makeGRPCCall<CreateFaqRequest, FaqServiceClient, CreateFaqResponse>(
			data,
			faqServiceClient,
			"createFaq"
		)
			.then((response: CreateFaqResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateFaq = (
	data: UpdateFaqRequest
): Promise<UpdateFaqResponse> => {
	return new Promise<UpdateFaqResponse>((resolve, reject) => {
		makeGRPCCall<UpdateFaqRequest, FaqServiceClient, UpdateFaqResponse>(
			data,
			faqServiceClient,
			"updateFaq"
		)
			.then((response: UpdateFaqResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readFaq = (data: ReadFaqRequest): Promise<ReadFaqResponse> => {
	return new Promise<ReadFaqResponse>((resolve, reject) => {
		makeGRPCCall<ReadFaqRequest, FaqServiceClient, ReadFaqResponse>(
			data,
			faqServiceClient,
			"readFaq"
		)
			.then((response: ReadFaqResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readFaqs = (data: ReadFaqsRequest): Promise<ReadFaqsResponse> => {
	return new Promise<ReadFaqsResponse>((resolve, reject) => {
		makeGRPCCall<ReadFaqsRequest, FaqServiceClient, ReadFaqsResponse>(
			data,
			faqServiceClient,
			"readFaqs"
		)
			.then((response: ReadFaqsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteFaq = (
	data: DeleteFaqRequest
): Promise<DeleteFaqResponse> => {
	return new Promise<DeleteFaqResponse>((resolve, reject) => {
		makeGRPCCall<DeleteFaqRequest, FaqServiceClient, DeleteFaqResponse>(
			data,
			faqServiceClient,
			"deleteFaq"
		)
			.then((response: DeleteFaqResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketMacro = (
	data: CreateTicketMacroRequest
): Promise<CreateTicketMacroResponse> => {
	return new Promise<CreateTicketMacroResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketMacroRequest,
			TicketMacroServiceClient,
			CreateTicketMacroResponse
		>(data, ticketMacroServiceClient, "createTicketMacro")
			.then((response: CreateTicketMacroResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketMacro = (
	data: UpdateTicketMacroRequest
): Promise<UpdateTicketMacroResponse> => {
	return new Promise<UpdateTicketMacroResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketMacroRequest,
			TicketMacroServiceClient,
			UpdateTicketMacroResponse
		>(data, ticketMacroServiceClient, "updateTicketMacro")
			.then((response: UpdateTicketMacroResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketMacro = (
	data: ReadTicketMacroRequest
): Promise<ReadTicketMacroResponse> => {
	return new Promise<ReadTicketMacroResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketMacroRequest,
			TicketMacroServiceClient,
			ReadTicketMacroResponse
		>(data, ticketMacroServiceClient, "readTicketMacro")
			.then((response: ReadTicketMacroResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketMacros = (
	data: ReadTicketMacrosRequest
): Promise<ReadTicketMacrosResponse> => {
	return new Promise<ReadTicketMacrosResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketMacrosRequest,
			TicketMacroServiceClient,
			ReadTicketMacrosResponse
		>(data, ticketMacroServiceClient, "readTicketMacros")
			.then((response: ReadTicketMacrosResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketMacro = (
	data: DeleteTicketMacroRequest
): Promise<DeleteTicketMacroResponse> => {
	return new Promise<DeleteTicketMacroResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketMacroRequest,
			TicketMacroServiceClient,
			DeleteTicketMacroResponse
		>(data, ticketMacroServiceClient, "deleteTicketMacro")
			.then((response: DeleteTicketMacroResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketScript = (
	data: CreateTicketScriptRequest
): Promise<CreateTicketScriptResponse> => {
	return new Promise<CreateTicketScriptResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketScriptRequest,
			TicketScriptServiceClient,
			CreateTicketScriptResponse
		>(data, ticketScriptServiceClient, "createTicketScript")
			.then((response: CreateTicketScriptResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketScript = (
	data: UpdateTicketScriptRequest
): Promise<UpdateTicketScriptResponse> => {
	return new Promise<UpdateTicketScriptResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketScriptRequest,
			TicketScriptServiceClient,
			UpdateTicketScriptResponse
		>(data, ticketScriptServiceClient, "updateTicketScript")
			.then((response: UpdateTicketScriptResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketScript = (
	data: ReadTicketScriptRequest
): Promise<ReadTicketScriptResponse> => {
	return new Promise<ReadTicketScriptResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketScriptRequest,
			TicketScriptServiceClient,
			ReadTicketScriptResponse
		>(data, ticketScriptServiceClient, "readTicketScript")
			.then((response: ReadTicketScriptResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketScripts = (
	data: ReadTicketScriptsRequest
): Promise<ReadTicketScriptsResponse> => {
	return new Promise<ReadTicketScriptsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketScriptsRequest,
			TicketScriptServiceClient,
			ReadTicketScriptsResponse
		>(data, ticketScriptServiceClient, "readTicketScripts")
			.then((response: ReadTicketScriptsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketScript = (
	data: DeleteTicketScriptRequest
): Promise<DeleteTicketScriptResponse> => {
	return new Promise<DeleteTicketScriptResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketScriptRequest,
			TicketScriptServiceClient,
			DeleteTicketScriptResponse
		>(data, ticketScriptServiceClient, "deleteTicketScript")
			.then((response: DeleteTicketScriptResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketCategory = (
	data: CreateTicketCategoryRequest
): Promise<CreateTicketCategoryResponse> => {
	return new Promise<CreateTicketCategoryResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketCategoryRequest,
			TicketCategoryServiceClient,
			CreateTicketCategoryResponse
		>(data, ticketCategoryServiceClient, "createTicketCategory")
			.then((response: CreateTicketCategoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketCategory = (
	data: UpdateTicketCategoryRequest
): Promise<UpdateTicketCategoryResponse> => {
	return new Promise<UpdateTicketCategoryResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketCategoryRequest,
			TicketCategoryServiceClient,
			UpdateTicketCategoryResponse
		>(data, ticketCategoryServiceClient, "updateTicketCategory")
			.then((response: UpdateTicketCategoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketCategory = (
	data: ReadTicketCategoryRequest
): Promise<ReadTicketCategoryResponse> => {
	return new Promise<ReadTicketCategoryResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketCategoryRequest,
			TicketCategoryServiceClient,
			ReadTicketCategoryResponse
		>(data, ticketCategoryServiceClient, "readTicketCategory")
			.then((response: ReadTicketCategoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketCategories = (
	data: ReadTicketCategoriesRequest
): Promise<ReadTicketCategoriesResponse> => {
	return new Promise<ReadTicketCategoriesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketCategoriesRequest,
			TicketCategoryServiceClient,
			ReadTicketCategoriesResponse
		>(data, ticketCategoryServiceClient, "readTicketCategories")
			.then((response: ReadTicketCategoriesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketCategory = (
	data: DeleteTicketCategoryRequest
): Promise<DeleteTicketCategoryResponse> => {
	return new Promise<DeleteTicketCategoryResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketCategoryRequest,
			TicketCategoryServiceClient,
			DeleteTicketCategoryResponse
		>(data, ticketCategoryServiceClient, "deleteTicketCategory")
			.then((response: DeleteTicketCategoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketTag = (
	data: CreateTicketTagRequest
): Promise<CreateTicketTagResponse> => {
	return new Promise<CreateTicketTagResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketTagRequest,
			TicketTagServiceClient,
			CreateTicketTagResponse
		>(data, ticketTagServiceClient, "createTicketTag")
			.then((response: CreateTicketTagResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketTag = (
	data: UpdateTicketTagRequest
): Promise<UpdateTicketTagResponse> => {
	return new Promise<UpdateTicketTagResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketTagRequest,
			TicketTagServiceClient,
			UpdateTicketTagResponse
		>(data, ticketTagServiceClient, "updateTicketTag")
			.then((response: UpdateTicketTagResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketTag = (
	data: ReadTicketTagRequest
): Promise<ReadTicketTagResponse> => {
	return new Promise<ReadTicketTagResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketTagRequest,
			TicketTagServiceClient,
			ReadTicketTagResponse
		>(data, ticketTagServiceClient, "readTicketTag")
			.then((response: ReadTicketTagResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketTags = (
	data: ReadTicketTagsRequest
): Promise<ReadTicketTagsResponse> => {
	return new Promise<ReadTicketTagsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketTagsRequest,
			TicketTagServiceClient,
			ReadTicketTagsResponse
		>(data, ticketTagServiceClient, "readTicketTags")
			.then((response: ReadTicketTagsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketTag = (
	data: DeleteTicketTagRequest
): Promise<DeleteTicketTagResponse> => {
	return new Promise<DeleteTicketTagResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketTagRequest,
			TicketTagServiceClient,
			DeleteTicketTagResponse
		>(data, ticketTagServiceClient, "deleteTicketTag")
			.then((response: DeleteTicketTagResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketRoom = (
	data: CreateTicketRoomRequest
): Promise<CreateTicketRoomResponse> => {
	return new Promise<CreateTicketRoomResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketRoomRequest,
			TicketRoomServiceClient,
			CreateTicketRoomResponse
		>(data, ticketRoomServiceClient, "createTicketRoom")
			.then((response: CreateTicketRoomResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketRoom = (
	data: UpdateTicketRoomRequest
): Promise<UpdateTicketRoomResponse> => {
	return new Promise<UpdateTicketRoomResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketRoomRequest,
			TicketRoomServiceClient,
			UpdateTicketRoomResponse
		>(data, ticketRoomServiceClient, "updateTicketRoom")
			.then((response: UpdateTicketRoomResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketRoom = (
	data: ReadTicketRoomRequest
): Promise<ReadTicketRoomResponse> => {
	return new Promise<ReadTicketRoomResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketRoomRequest,
			TicketRoomServiceClient,
			ReadTicketRoomResponse
		>(data, ticketRoomServiceClient, "readTicketRoom")
			.then((response: ReadTicketRoomResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketRooms = (
	data: ReadTicketRoomsRequest
): Promise<ReadTicketRoomsResponse> => {
	return new Promise<ReadTicketRoomsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketRoomsRequest,
			TicketRoomServiceClient,
			ReadTicketRoomsResponse
		>(data, ticketRoomServiceClient, "readTicketRooms")
			.then((response: ReadTicketRoomsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketRoom = (
	data: DeleteTicketRoomRequest
): Promise<DeleteTicketRoomResponse> => {
	return new Promise<DeleteTicketRoomResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketRoomRequest,
			TicketRoomServiceClient,
			DeleteTicketRoomResponse
		>(data, ticketRoomServiceClient, "deleteTicketRoom")
			.then((response: DeleteTicketRoomResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketLogs = (
	data: ReadTicketLogsRequest
): Promise<ReadTicketLogsResponse> => {
	return new Promise<ReadTicketLogsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketLogsRequest,
			TicketLogServiceClient,
			ReadTicketLogsResponse
		>(data, ticketLogServiceClient, "readTicketLogs")
			.then((response: ReadTicketLogsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketCustomFieldDefinition = (
	data: CreateTicketCustomFieldDefinitionRequest
): Promise<CreateTicketCustomFieldDefinitionResponse> => {
	return new Promise<CreateTicketCustomFieldDefinitionResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				CreateTicketCustomFieldDefinitionRequest,
				TicketCustomFieldDefinitionServiceClient,
				CreateTicketCustomFieldDefinitionResponse
			>(
				data,
				ticketCustomFieldDefinitionServiceClient,
				"createTicketCustomFieldDefinition"
			)
				.then((response: CreateTicketCustomFieldDefinitionResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const updateTicketCustomFieldDefinition = (
	data: UpdateTicketCustomFieldDefinitionRequest
): Promise<UpdateTicketCustomFieldDefinitionResponse> => {
	return new Promise<UpdateTicketCustomFieldDefinitionResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				UpdateTicketCustomFieldDefinitionRequest,
				TicketCustomFieldDefinitionServiceClient,
				UpdateTicketCustomFieldDefinitionResponse
			>(
				data,
				ticketCustomFieldDefinitionServiceClient,
				"updateTicketCustomFieldDefinition"
			)
				.then((response: UpdateTicketCustomFieldDefinitionResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const readTicketCustomFieldDefinitions = (
	data: ReadTicketCustomFieldDefinitionsRequest
): Promise<ReadTicketCustomFieldDefinitionsResponse> => {
	return new Promise<ReadTicketCustomFieldDefinitionsResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				ReadTicketCustomFieldDefinitionsRequest,
				TicketCustomFieldDefinitionServiceClient,
				ReadTicketCustomFieldDefinitionsResponse
			>(
				data,
				ticketCustomFieldDefinitionServiceClient,
				"readTicketCustomFieldDefinitions"
			)
				.then((response: ReadTicketCustomFieldDefinitionsResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const deleteTicketCustomFieldDefinition = (
	data: DeleteTicketCustomFieldDefinitionRequest
): Promise<DeleteTicketCustomFieldDefinitionResponse> => {
	return new Promise<DeleteTicketCustomFieldDefinitionResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				DeleteTicketCustomFieldDefinitionRequest,
				TicketCustomFieldDefinitionServiceClient,
				DeleteTicketCustomFieldDefinitionResponse
			>(
				data,
				ticketCustomFieldDefinitionServiceClient,
				"deleteTicketCustomFieldDefinition"
			)
				.then((response: DeleteTicketCustomFieldDefinitionResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const createTicketNote = (
	data: CreateTicketNoteRequest
): Promise<CreateTicketNoteResponse> => {
	return new Promise<CreateTicketNoteResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketNoteRequest,
			TicketNoteServiceClient,
			CreateTicketNoteResponse
		>(data, ticketNoteServiceClient, "createTicketNote")
			.then((response: CreateTicketNoteResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicketNote = (
	data: UpdateTicketNoteRequest
): Promise<UpdateTicketNoteResponse> => {
	return new Promise<UpdateTicketNoteResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketNoteRequest,
			TicketNoteServiceClient,
			UpdateTicketNoteResponse
		>(data, ticketNoteServiceClient, "updateTicketNote")
			.then((response: UpdateTicketNoteResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketNote = (
	data: ReadTicketNoteRequest
): Promise<ReadTicketNoteResponse> => {
	return new Promise<ReadTicketNoteResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketNoteRequest,
			TicketNoteServiceClient,
			ReadTicketNoteResponse
		>(data, ticketNoteServiceClient, "readTicketNote")
			.then((response: ReadTicketNoteResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketNotes = (
	data: ReadTicketNotesRequest
): Promise<ReadTicketNotesResponse> => {
	return new Promise<ReadTicketNotesResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketNotesRequest,
			TicketNoteServiceClient,
			ReadTicketNotesResponse
		>(data, ticketNoteServiceClient, "readTicketNotes")
			.then((response: ReadTicketNotesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteTicketNote = (
	data: DeleteTicketNoteRequest
): Promise<DeleteTicketNoteResponse> => {
	return new Promise<DeleteTicketNoteResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteTicketNoteRequest,
			TicketNoteServiceClient,
			DeleteTicketNoteResponse
		>(data, ticketNoteServiceClient, "deleteTicketNote")
			.then((response: DeleteTicketNoteResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createAttachment = (
	data: CreateAttachmentRequest
): Promise<CreateAttachmentResponse> => {
	return new Promise<CreateAttachmentResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateAttachmentRequest,
			AttachmentServiceClient,
			CreateAttachmentResponse
		>(data, attachmentServiceClient, "createAttachment")
			.then((response: CreateAttachmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readAttachment = (
	data: ReadAttachmentRequest
): Promise<ReadAttachmentResponse> => {
	return new Promise<ReadAttachmentResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadAttachmentRequest,
			AttachmentServiceClient,
			ReadAttachmentResponse
		>(data, attachmentServiceClient, "readAttachment")
			.then((response: ReadAttachmentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readAttachments = (
	data: ReadAttachmentsRequest
): Promise<ReadAttachmentsResponse> => {
	return new Promise<ReadAttachmentsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadAttachmentsRequest,
			AttachmentServiceClient,
			ReadAttachmentsResponse
		>(data, attachmentServiceClient, "readAttachments")
			.then((response: ReadAttachmentsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getAttachmentPresignedURL = (
	data: GetAttachmentPresignedURLRequest
): Promise<GetAttachmentPresignedURLResponse> => {
	return new Promise<GetAttachmentPresignedURLResponse>((resolve, reject) => {
		makeGRPCCall<
			GetAttachmentPresignedURLRequest,
			AttachmentServiceClient,
			GetAttachmentPresignedURLResponse
		>(data, attachmentServiceClient, "getAttachmentPresignedURL")
			.then((response: GetAttachmentPresignedURLResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketsInRoom = (
	data: ReadTicketsInRoomRequest
): Promise<ReadTicketsInRoomResponse> => {
	return new Promise<ReadTicketsInRoomResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketsInRoomRequest,
			TicketServiceClient,
			ReadTicketsInRoomResponse
		>(data, ticketServiceClient, "readTicketsInRoom")
			.then((response: ReadTicketsInRoomResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsCategory = (
	data: ChangeTicketsCategoryRequest
): Promise<ChangeTicketsCategoryResponse> => {
	return new Promise<ChangeTicketsCategoryResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsCategoryRequest,
			TicketServiceClient,
			ChangeTicketsCategoryResponse
		>(data, ticketServiceClient, "changeTicketsCategory")
			.then((response: ChangeTicketsCategoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsPriority = (
	data: ChangeTicketsPriorityRequest
): Promise<ChangeTicketsPriorityResponse> => {
	return new Promise<ChangeTicketsPriorityResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsPriorityRequest,
			TicketServiceClient,
			ChangeTicketsPriorityResponse
		>(data, ticketServiceClient, "changeTicketsPriority")
			.then((response: ChangeTicketsPriorityResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsStatus = (
	data: ChangeTicketsStatusRequest
): Promise<ChangeTicketsStatusResponse> => {
	return new Promise<ChangeTicketsStatusResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsStatusRequest,
			TicketServiceClient,
			ChangeTicketsStatusResponse
		>(data, ticketServiceClient, "changeTicketsStatus")
			.then((response: ChangeTicketsStatusResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsTags = (
	data: ChangeTicketsTagsRequest
): Promise<ChangeTicketsTagsResponse> => {
	return new Promise<ChangeTicketsTagsResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsTagsRequest,
			TicketServiceClient,
			ChangeTicketsTagsResponse
		>(data, ticketServiceClient, "changeTicketsTags")
			.then((response: ChangeTicketsTagsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const markTicketsAsRead = (
	data: MarkTicketsAsReadRequest
): Promise<MarkTicketsAsReadResponse> => {
	return new Promise<MarkTicketsAsReadResponse>((resolve, reject) => {
		makeGRPCCall<
			MarkTicketsAsReadRequest,
			TicketServiceClient,
			MarkTicketsAsReadResponse
		>(data, ticketServiceClient, "markTicketsAsRead")
			.then((response: MarkTicketsAsReadResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const markTicketsAsUnread = (
	data: MarkTicketsAsUnreadRequest
): Promise<MarkTicketsAsUnreadResponse> => {
	return new Promise<MarkTicketsAsUnreadResponse>((resolve, reject) => {
		makeGRPCCall<
			MarkTicketsAsUnreadRequest,
			TicketServiceClient,
			MarkTicketsAsUnreadResponse
		>(data, ticketServiceClient, "markTicketsAsUnread")
			.then((response: MarkTicketsAsUnreadResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const markTicketsAsSpam = (
	data: MarkTicketsAsSpamRequest
): Promise<MarkTicketsAsSpamResponse> => {
	return new Promise<MarkTicketsAsSpamResponse>((resolve, reject) => {
		makeGRPCCall<
			MarkTicketsAsSpamRequest,
			TicketServiceClient,
			MarkTicketsAsSpamResponse
		>(data, ticketServiceClient, "markTicketsAsSpam")
			.then((response: MarkTicketsAsSpamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const markTicketsAsNotSpam = (
	data: MarkTicketsAsNotSpamRequest
): Promise<MarkTicketsAsNotSpamResponse> => {
	return new Promise<MarkTicketsAsNotSpamResponse>((resolve, reject) => {
		makeGRPCCall<
			MarkTicketsAsNotSpamRequest,
			TicketServiceClient,
			MarkTicketsAsNotSpamResponse
		>(data, ticketServiceClient, "markTicketsAsNotSpam")
			.then((response: MarkTicketsAsNotSpamResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const archiveTickets = (
	data: ArchiveTicketsRequest
): Promise<ArchiveTicketsResponse> => {
	return new Promise<ArchiveTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ArchiveTicketsRequest,
			TicketServiceClient,
			ArchiveTicketsResponse
		>(data, ticketServiceClient, "archiveTickets")
			.then((response: ArchiveTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const unarchiveTickets = (
	data: UnarchiveTicketsRequest
): Promise<UnarchiveTicketsResponse> => {
	return new Promise<UnarchiveTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			UnarchiveTicketsRequest,
			TicketServiceClient,
			UnarchiveTicketsResponse
		>(data, ticketServiceClient, "unarchiveTickets")
			.then((response: UnarchiveTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsAssignedAgent = (
	data: ChangeTicketsAssignedAgentRequest
): Promise<ChangeTicketsAssignedAgentResponse> => {
	return new Promise<ChangeTicketsAssignedAgentResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsAssignedAgentRequest,
			TicketServiceClient,
			ChangeTicketsAssignedAgentResponse
		>(data, ticketServiceClient, "changeTicketsAssignedAgent")
			.then((response: ChangeTicketsAssignedAgentResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTickets = (
	data: ReadTicketsRequest
): Promise<ReadTicketsResponse> => {
	return new Promise<ReadTicketsResponse>((resolve, reject) => {
		makeGRPCCall<ReadTicketsRequest, TicketServiceClient, ReadTicketsResponse>(
			data,
			ticketServiceClient,
			"readTickets"
		)
			.then((response: ReadTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readMyAssignedTickets = (
	data: ReadMyAssignedTicketsRequest
): Promise<ReadMyAssignedTicketsResponse> => {
	return new Promise<ReadMyAssignedTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadMyAssignedTicketsRequest,
			TicketServiceClient,
			ReadMyAssignedTicketsResponse
		>(data, ticketServiceClient, "readMyAssignedTickets")
			.then((response: ReadMyAssignedTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

// export const readMentionedTickets = (
// 	data: ReadMentionedTicketsRequest
// ): Promise<ReadMentionedTicketsResponse> => {
// 	return new Promise<ReadMentionedTicketsResponse>((resolve, reject) => {
// 		makeGRPCCall<
// 			ReadMentionedTicketsRequest,
// 			TicketServiceClient,
// 			ReadMentionedTicketsResponse
// 		>(data, ticketServiceClient, "readMentionedTickets")
// 			.then((response: ReadMentionedTicketsResponse) => {
// 				resolve(response);
// 			})
// 			.catch((error: RpcError) => {
// 				reject(error);
// 			});
// 	});
// };

export const readUnassignedTickets = (
	data: ReadUnassignedTicketsRequest
): Promise<ReadUnassignedTicketsResponse> => {
	return new Promise<ReadUnassignedTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadUnassignedTicketsRequest,
			TicketServiceClient,
			ReadUnassignedTicketsResponse
		>(data, ticketServiceClient, "readUnassignedTickets")
			.then((response: ReadUnassignedTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readAllArchivedTickets = (
	data: ReadArchivedTicketsRequest
): Promise<ReadArchivedTicketsResponse> => {
	return new Promise<ReadArchivedTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadArchivedTicketsRequest,
			TicketServiceClient,
			ReadArchivedTicketsResponse
		>(data, ticketServiceClient, "readAllArchivedTickets")
			.then((response: ReadArchivedTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readMyArchivedTickets = (
	data: ReadArchivedTicketsRequest
): Promise<ReadArchivedTicketsResponse> => {
	return new Promise<ReadArchivedTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadArchivedTicketsRequest,
			TicketServiceClient,
			ReadArchivedTicketsResponse
		>(data, ticketServiceClient, "readMyArchivedTickets")
			.then((response: ReadArchivedTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readSpamTickets = (
	data: ReadSpamTicketsRequest
): Promise<ReadSpamTicketsResponse> => {
	return new Promise<ReadSpamTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadSpamTicketsRequest,
			TicketServiceClient,
			ReadSpamTicketsResponse
		>(data, ticketServiceClient, "readSpamTickets")
			.then((response: ReadSpamTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createSLAPolicy = (
	data: CreateSLAPolicyRequest
): Promise<CreateSLAPolicyResponse> => {
	return new Promise<CreateSLAPolicyResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateSLAPolicyRequest,
			SLAServiceClient,
			CreateSLAPolicyResponse
		>(data, slaServiceClient, "createSLAPolicy")
			.then((response: CreateSLAPolicyResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateSLAPolicy = (
	data: UpdateSLAPolicyRequest
): Promise<UpdateSLAPolicyResponse> => {
	return new Promise<UpdateSLAPolicyResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateSLAPolicyRequest,
			SLAServiceClient,
			UpdateSLAPolicyResponse
		>(data, slaServiceClient, "updateSLAPolicy")
			.then((response: UpdateSLAPolicyResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteSLAPolicy = (
	data: DeleteSLAPolicyRequest
): Promise<DeleteSLAPolicyResponse> => {
	return new Promise<DeleteSLAPolicyResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteSLAPolicyRequest,
			SLAServiceClient,
			DeleteSLAPolicyResponse
		>(data, slaServiceClient, "deleteSLAPolicy")
			.then((response: DeleteSLAPolicyResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const listSLAPolicies = (
	data: ListSLAPoliciesRequest
): Promise<ListSLAPoliciesResponse> => {
	return new Promise<ListSLAPoliciesResponse>((resolve, reject) => {
		makeGRPCCall<
			ListSLAPoliciesRequest,
			SLAServiceClient,
			ListSLAPoliciesResponse
		>(data, slaServiceClient, "listSLAPolicies")
			.then((response: ListSLAPoliciesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getTicketSLA = (
	data: GetTicketSLARequest
): Promise<GetTicketSLAResponse> => {
	return new Promise<GetTicketSLAResponse>((resolve, reject) => {
		makeGRPCCall<GetTicketSLARequest, SLAServiceClient, GetTicketSLAResponse>(
			data,
			slaServiceClient,
			"getTicketSla"
		)
			.then((response: GetTicketSLAResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createThread = (
	data: CreateThreadRequest
): Promise<CreateThreadResponse> => {
	return new Promise<CreateThreadResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateThreadRequest,
			ThreadServiceClient,
			CreateThreadResponse
		>(data, threadServiceClient, "createThread")
			.then((response: CreateThreadResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createEscalationThread = (
	data: CreateEscalationThreadRequest
): Promise<CreateEscalationThreadResponse> => {
	return new Promise<CreateEscalationThreadResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateEscalationThreadRequest,
			EscalationThreadServiceClient,
			CreateEscalationThreadResponse
		>(data, escalationThreadServiceClient, "createEscalationThread")
			.then((response: CreateEscalationThreadResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicket = (
	data: CreateTicketRequest
): Promise<CreateTicketResponse> => {
	return new Promise<CreateTicketResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateTicketRequest,
			TicketServiceClient,
			CreateTicketResponse
		>(data, ticketServiceClient, "createTicket")
			.then((response: CreateTicketResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateTicket = (
	data: UpdateTicketRequest
): Promise<UpdateTicketResponse> => {
	return new Promise<UpdateTicketResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateTicketRequest,
			TicketServiceClient,
			UpdateTicketResponse
		>(data, ticketServiceClient, "updateTicket")
			.then((response: UpdateTicketResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const changeTicketsFollowers = (
	data: ChangeTicketsFollowersRequest
): Promise<ChangeTicketsFollowersResponse> => {
	return new Promise<ChangeTicketsFollowersResponse>((resolve, reject) => {
		makeGRPCCall<
			ChangeTicketsFollowersRequest,
			TicketServiceClient,
			ChangeTicketsFollowersResponse
		>(data, ticketServiceClient, "changeTicketsFollowers")
			.then((response: ChangeTicketsFollowersResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const mergeTickets = (
	data: MergeTicketsRequest
): Promise<MergeTicketsResponse> => {
	return new Promise<MergeTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			MergeTicketsRequest,
			TicketServiceClient,
			MergeTicketsResponse
		>(data, ticketServiceClient, "mergeTickets")
			.then((response: MergeTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getPotentialDuplicates = (
	data: GetPotentialDuplicatesRequest
): Promise<GetPotentialDuplicatesResponse> => {
	return new Promise<GetPotentialDuplicatesResponse>((resolve, reject) => {
		makeGRPCCall<
			GetPotentialDuplicatesRequest,
			TicketServiceClient,
			GetPotentialDuplicatesResponse
		>(data, ticketServiceClient, "getPotentialDuplicates")
			.then((response: GetPotentialDuplicatesResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const reviewDuplicate = (
	data: ReviewDuplicateRequest
): Promise<ReviewDuplicateResponse> => {
	return new Promise<ReviewDuplicateResponse>((resolve, reject) => {
		makeGRPCCall<
			ReviewDuplicateRequest,
			TicketServiceClient,
			ReviewDuplicateResponse
		>(data, ticketServiceClient, "reviewDuplicate")
			.then((response: ReviewDuplicateResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getMergeHistory = (
	data: GetMergeHistoryRequest
): Promise<GetMergeHistoryResponse> => {
	return new Promise<GetMergeHistoryResponse>((resolve, reject) => {
		makeGRPCCall<
			GetMergeHistoryRequest,
			TicketServiceClient,
			GetMergeHistoryResponse
		>(data, ticketServiceClient, "getMergeHistory")
			.then((response: GetMergeHistoryResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getMergedTickets = (
	data: GetMergedTicketsRequest
): Promise<GetMergedTicketsResponse> => {
	return new Promise<GetMergedTicketsResponse>((resolve, reject) => {
		makeGRPCCall<
			GetMergedTicketsRequest,
			TicketServiceClient,
			GetMergedTicketsResponse
		>(data, ticketServiceClient, "getMergedTickets")
			.then((response: GetMergedTicketsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getMasterTicket = (
	data: GetMasterTicketRequest
): Promise<GetMasterTicketResponse> => {
	return new Promise<GetMasterTicketResponse>((resolve, reject) => {
		makeGRPCCall<
			GetMasterTicketRequest,
			TicketServiceClient,
			GetMasterTicketResponse
		>(data, ticketServiceClient, "getMasterTicket")
			.then((response: GetMasterTicketResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicket = (
	data: ReadTicketRequest
): Promise<ReadTicketResponse> => {
	return new Promise<ReadTicketResponse>((resolve, reject) => {
		makeGRPCCall<ReadTicketRequest, TicketServiceClient, ReadTicketResponse>(
			data,
			ticketServiceClient,
			"readTicket"
		)
			.then((response: ReadTicketResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getKPIMetrics = (
	data: GetKPIMetricsRequest
): Promise<KPIMetrics> => {
	return new Promise<KPIMetrics>((resolve, reject) => {
		makeGRPCCall<
			GetKPIMetricsRequest,
			TicketAnalyticsServiceClient,
			KPIMetrics
		>(data, ticketAnalyticsServiceClient, "getKPIMetrics")
			.then((response: KPIMetrics) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getTicketVolume = (
	data: GetTicketVolumeRequest
): Promise<TicketVolumeResponse> => {
	return new Promise<TicketVolumeResponse>((resolve, reject) => {
		makeGRPCCall<
			GetTicketVolumeRequest,
			TicketAnalyticsServiceClient,
			TicketVolumeResponse
		>(data, ticketAnalyticsServiceClient, "getTicketVolume")
			.then((response: TicketVolumeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getChannelDistribution = (
	data: GetChannelDistributionRequest
): Promise<ChannelDistributionResponse> => {
	return new Promise<ChannelDistributionResponse>((resolve, reject) => {
		makeGRPCCall<
			GetChannelDistributionRequest,
			TicketAnalyticsServiceClient,
			ChannelDistributionResponse
		>(data, ticketAnalyticsServiceClient, "getChannelDistribution")
			.then((response: ChannelDistributionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getIntegrationDistribution = (
	data: GetIntegrationDistributionRequest
): Promise<IntegrationDistributionResponse> => {
	return new Promise<IntegrationDistributionResponse>((resolve, reject) => {
		makeGRPCCall<
			GetIntegrationDistributionRequest,
			TicketAnalyticsServiceClient,
			IntegrationDistributionResponse
		>(data, ticketAnalyticsServiceClient, "getIntegrationDistribution")
			.then((response: IntegrationDistributionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getStatusDistribution = (
	data: GetStatusDistributionRequest
): Promise<StatusDistributionResponse> => {
	return new Promise<StatusDistributionResponse>((resolve, reject) => {
		makeGRPCCall<
			GetStatusDistributionRequest,
			TicketAnalyticsServiceClient,
			StatusDistributionResponse
		>(data, ticketAnalyticsServiceClient, "getStatusDistribution")
			.then((response: StatusDistributionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getPriorityDistribution = (
	data: GetPriorityDistributionRequest
): Promise<PriorityDistributionResponse> => {
	return new Promise<PriorityDistributionResponse>((resolve, reject) => {
		makeGRPCCall<
			GetPriorityDistributionRequest,
			TicketAnalyticsServiceClient,
			PriorityDistributionResponse
		>(data, ticketAnalyticsServiceClient, "getPriorityDistribution")
			.then((response: PriorityDistributionResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getAgentPerformance = (
	data: GetAgentPerformanceRequest
): Promise<AgentPerformanceResponse> => {
	return new Promise<AgentPerformanceResponse>((resolve, reject) => {
		makeGRPCCall<
			GetAgentPerformanceRequest,
			TicketAnalyticsServiceClient,
			AgentPerformanceResponse
		>(data, ticketAnalyticsServiceClient, "getAgentPerformance")
			.then((response: AgentPerformanceResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getSLACompliance = (
	data: GetSLAComplianceRequest
): Promise<SLAComplianceResponse> => {
	return new Promise<SLAComplianceResponse>((resolve, reject) => {
		makeGRPCCall<
			GetSLAComplianceRequest,
			TicketAnalyticsServiceClient,
			SLAComplianceResponse
		>(data, ticketAnalyticsServiceClient, "getSLACompliance")
			.then((response: SLAComplianceResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getActivityHeatmap = (
	data: GetActivityHeatmapRequest
): Promise<ActivityHeatmapResponse> => {
	return new Promise<ActivityHeatmapResponse>((resolve, reject) => {
		makeGRPCCall<
			GetActivityHeatmapRequest,
			TicketAnalyticsServiceClient,
			ActivityHeatmapResponse
		>(data, ticketAnalyticsServiceClient, "getActivityHeatmap")
			.then((response: ActivityHeatmapResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getResolutionTimeAnalytics = (
	data: GetResolutionTimeAnalyticsRequest
): Promise<ResolutionTimeResponse> => {
	return new Promise<ResolutionTimeResponse>((resolve, reject) => {
		makeGRPCCall<
			GetResolutionTimeAnalyticsRequest,
			TicketAnalyticsServiceClient,
			ResolutionTimeResponse
		>(data, ticketAnalyticsServiceClient, "getResolutionTimeAnalytics")
			.then((response: ResolutionTimeResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getEscalationMetrics = (
	data: GetEscalationMetricsRequest
): Promise<EscalationMetrics> => {
	return new Promise<EscalationMetrics>((resolve, reject) => {
		makeGRPCCall<
			GetEscalationMetricsRequest,
			TicketAnalyticsServiceClient,
			EscalationMetrics
		>(data, ticketAnalyticsServiceClient, "getEscalationMetrics")
			.then((response: EscalationMetrics) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getLiveTicketTracking = (
	data: GetLiveTicketTrackingRequest
): Promise<LiveTicketTrackingResponse> => {
	return new Promise<LiveTicketTrackingResponse>((resolve, reject) => {
		makeGRPCCall<
			GetLiveTicketTrackingRequest,
			TicketAnalyticsServiceClient,
			LiveTicketTrackingResponse
		>(data, ticketAnalyticsServiceClient, "getLiveTicketTracking")
			.then((response: LiveTicketTrackingResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getTrendAnalysis = (
	data: GetTrendAnalysisRequest
): Promise<TrendAnalysisResponse> => {
	return new Promise<TrendAnalysisResponse>((resolve, reject) => {
		makeGRPCCall<
			GetTrendAnalysisRequest,
			TicketAnalyticsServiceClient,
			TrendAnalysisResponse
		>(data, ticketAnalyticsServiceClient, "getTrendAnalysis")
			.then((response: TrendAnalysisResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getSentimentAnalytics = (
	data: GetSentimentAnalyticsRequest
): Promise<SentimentAnalyticsResponse> => {
	return new Promise<SentimentAnalyticsResponse>((resolve, reject) => {
		makeGRPCCall<
			GetSentimentAnalyticsRequest,
			TicketAnalyticsServiceClient,
			SentimentAnalyticsResponse
		>(data, ticketAnalyticsServiceClient, "getSentimentAnalytics")
			.then((response: SentimentAnalyticsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createDashboard = (
	data: CreateDashboardRequest
): Promise<DashboardResponse> => {
	return new Promise<DashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			CreateDashboardRequest,
			TicketDashboardServiceClient,
			DashboardResponse
		>(data, ticketDashboardServiceClient, "createDashboard")
			.then((response: DashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getDashboard = (
	data: GetDashboardRequest
): Promise<DashboardResponse> => {
	return new Promise<DashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			GetDashboardRequest,
			TicketDashboardServiceClient,
			DashboardResponse
		>(data, ticketDashboardServiceClient, "getDashboard")
			.then((response: DashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const updateDashboard = (
	data: UpdateDashboardRequest
): Promise<DashboardResponse> => {
	return new Promise<DashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			UpdateDashboardRequest,
			TicketDashboardServiceClient,
			DashboardResponse
		>(data, ticketDashboardServiceClient, "updateDashboard")
			.then((response: DashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const deleteDashboard = (
	data: DeleteDashboardRequest
): Promise<DeleteDashboardResponse> => {
	return new Promise<DeleteDashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			DeleteDashboardRequest,
			TicketDashboardServiceClient,
			DeleteDashboardResponse
		>(data, ticketDashboardServiceClient, "deleteDashboard")
			.then((response: DeleteDashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const listDashboards = (
	data: ListDashboardsRequest
): Promise<ListDashboardsResponse> => {
	return new Promise<ListDashboardsResponse>((resolve, reject) => {
		makeGRPCCall<
			ListDashboardsRequest,
			TicketDashboardServiceClient,
			ListDashboardsResponse
		>(data, ticketDashboardServiceClient, "listDashboards")
			.then((response: ListDashboardsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const duplicateDashboard = (
	data: DuplicateDashboardRequest
): Promise<DashboardResponse> => {
	return new Promise<DashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			DuplicateDashboardRequest,
			TicketDashboardServiceClient,
			DashboardResponse
		>(data, ticketDashboardServiceClient, "duplicateDashboard")
			.then((response: DashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const getAvailableMetrics = (
	data: GetAvailableMetricsRequest
): Promise<GetAvailableMetricsResponse> => {
	return new Promise<GetAvailableMetricsResponse>((resolve, reject) => {
		makeGRPCCall<
			GetAvailableMetricsRequest,
			TicketDashboardServiceClient,
			GetAvailableMetricsResponse
		>(data, ticketDashboardServiceClient, "getAvailableMetrics")
			.then((response: GetAvailableMetricsResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const validateDashboard = (
	data: ValidateDashboardRequest
): Promise<ValidateDashboardResponse> => {
	return new Promise<ValidateDashboardResponse>((resolve, reject) => {
		makeGRPCCall<
			ValidateDashboardRequest,
			TicketDashboardServiceClient,
			ValidateDashboardResponse
		>(data, ticketDashboardServiceClient, "validateDashboard")
			.then((response: ValidateDashboardResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const validateMetricConfig = (
	data: ValidateMetricConfigRequest
): Promise<ValidateMetricConfigResponse> => {
	return new Promise<ValidateMetricConfigResponse>((resolve, reject) => {
		makeGRPCCall<
			ValidateMetricConfigRequest,
			TicketDashboardServiceClient,
			ValidateMetricConfigResponse
		>(data, ticketDashboardServiceClient, "validateMetricConfig")
			.then((response: ValidateMetricConfigResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const createTicketAssignmentPolicy = (
	data: CreateTicketAssignmentPolicyRequest
): Promise<CreateTicketAssignmentPolicyResponse> => {
	return new Promise<CreateTicketAssignmentPolicyResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				CreateTicketAssignmentPolicyRequest,
				TicketAssignmentPolicyServiceClient,
				CreateTicketAssignmentPolicyResponse
			>(
				data,
				ticketAssignmentPolicyServiceClient,
				"createTicketAssignmentPolicy"
			)
				.then((response: CreateTicketAssignmentPolicyResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const updateTicketAssignmentPolicy = (
	data: UpdateTicketAssignmentPolicyRequest
): Promise<UpdateTicketAssignmentPolicyResponse> => {
	return new Promise<UpdateTicketAssignmentPolicyResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				UpdateTicketAssignmentPolicyRequest,
				TicketAssignmentPolicyServiceClient,
				UpdateTicketAssignmentPolicyResponse
			>(
				data,
				ticketAssignmentPolicyServiceClient,
				"updateTicketAssignmentPolicy"
			)
				.then((response: UpdateTicketAssignmentPolicyResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const deleteTicketAssignmentPolicy = (
	data: DeleteTicketAssignmentPolicyRequest
): Promise<DeleteTicketAssignmentPolicyResponse> => {
	return new Promise<DeleteTicketAssignmentPolicyResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				DeleteTicketAssignmentPolicyRequest,
				TicketAssignmentPolicyServiceClient,
				DeleteTicketAssignmentPolicyResponse
			>(
				data,
				ticketAssignmentPolicyServiceClient,
				"deleteTicketAssignmentPolicy"
			)
				.then((response: DeleteTicketAssignmentPolicyResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const readTicketAssignmentPolicy = (
	data: ReadTicketAssignmentPolicyRequest
): Promise<ReadTicketAssignmentPolicyResponse> => {
	return new Promise<ReadTicketAssignmentPolicyResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadTicketAssignmentPolicyRequest,
			TicketAssignmentPolicyServiceClient,
			ReadTicketAssignmentPolicyResponse
		>(data, ticketAssignmentPolicyServiceClient, "readTicketAssignmentPolicy")
			.then((response: ReadTicketAssignmentPolicyResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};

export const readTicketAssignmentPolicies = (
	data: ReadTicketAssignmentPoliciesRequest
): Promise<ReadTicketAssignmentPoliciesResponse> => {
	return new Promise<ReadTicketAssignmentPoliciesResponse>(
		(resolve, reject) => {
			makeGRPCCall<
				ReadTicketAssignmentPoliciesRequest,
				TicketAssignmentPolicyServiceClient,
				ReadTicketAssignmentPoliciesResponse
			>(
				data,
				ticketAssignmentPolicyServiceClient,
				"readTicketAssignmentPolicies"
			)
				.then((response: ReadTicketAssignmentPoliciesResponse) => {
					resolve(response);
				})
				.catch((error: RpcError) => {
					reject(error);
				});
		}
	);
};

export const readOwnerAssignmentPolicy = (
	data: ReadOwnerAssignmentPolicyRequest
): Promise<ReadOwnerAssignmentPolicyResponse> => {
	return new Promise<ReadOwnerAssignmentPolicyResponse>((resolve, reject) => {
		makeGRPCCall<
			ReadOwnerAssignmentPolicyRequest,
			TicketAssignmentPolicyServiceClient,
			ReadOwnerAssignmentPolicyResponse
		>(data, ticketAssignmentPolicyServiceClient, "readOwnerAssignmentPolicy")
			.then((response: ReadOwnerAssignmentPolicyResponse) => {
				resolve(response);
			})
			.catch((error: RpcError) => {
				reject(error);
			});
	});
};
