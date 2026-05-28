import { RpcError } from "grpc-web";

import {
  flowServiceClient,
  leadServiceClient,
  pipelineServiceClient,
  pipelineStageServiceClient,
  stageLabelServiceClient,
} from "../globalClient/GlobalClients";
import { makeGRPCCall } from "./rpcUtils";
import type {
  CreateCustomFieldRequest,
  CreateCustomFieldResponse,
  ReadCustomFieldRequest,
  ReadCustomFieldResponse,
  ReadLeadKeysRequest,
  ReadLeadKeysResponse,
  UpdateCustomFieldRequest,
  UpdateCustomFieldResponse,
  DeleteCustomFieldRequest,
  DeleteCustomFieldResponse,
  CreateLeadRequest,
  CreateLeadResponse,
  ReadLeadRequest,
  ReadLeadResponse,
  ListLeadRequest,
  ListLeadResponse,
  BulkLeadsImportRequest,
  BulkLeadsImportResponse,
} from "@/pb/lead";
import type { LeadServiceClient } from "@/pb/lead.client";
import {
  type CreatePipelineRequest,
  type CreatePipelineResponse,
  type ReadPipelinesRequest,
  type ReadPipelinesResponse,
  type ReadAppMetaDataRequest,
  type ReadAppMetaDataResponse,
  CreateFlowRequest,
  type CreateFlowResponse,
  type UpdateFlowRequest,
  type UpdateFlowResponse,
  type ReadFlowRequest,
  type ReadFlowResponse,
  type DeleteFlowRequest,
  type DeleteFlowResponse,
  type ReadPipelineStageRequest,
  type ReadPipelineStageResponse,
  type CreateStageLabelRequest,
  type CreateStageLabelResponse,
  type ReadStageLabelsRequest,
  type ReadStageLabelsResponse,
} from "@/pb/pipelines";
import type {
  PipelineServiceClient,
  FlowServiceClient,
  PipelineStageServiceClient,
  StageLabelServiceClient,
} from "@/pb/pipelines.client";

export const createPipeline = async (
  data: CreatePipelineRequest,
): Promise<CreatePipelineResponse> => {
  return new Promise<CreatePipelineResponse>((resolve, reject) => {
    makeGRPCCall<
      CreatePipelineRequest,
      PipelineServiceClient,
      CreatePipelineResponse
    >(data, pipelineServiceClient, "createPipeline")
      .then((response: CreatePipelineResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readPipelines = async (
  data: ReadPipelinesRequest,
): Promise<ReadPipelinesResponse> => {
  return new Promise<ReadPipelinesResponse>((resolve, reject) => {
    makeGRPCCall<
      ReadPipelinesRequest,
      PipelineServiceClient,
      ReadPipelinesResponse
    >(data, pipelineServiceClient, "readPipelines")
      .then((response: ReadPipelinesResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readUserPipelines = async (
  data: ReadAppMetaDataRequest,
): Promise<ReadAppMetaDataResponse> => {
  return new Promise<ReadAppMetaDataResponse>((resolve, reject) => {
    makeGRPCCall<
      ReadAppMetaDataRequest,
      PipelineServiceClient,
      ReadAppMetaDataResponse
    >(data, pipelineServiceClient, "readAppMetaData")
      .then((response: ReadAppMetaDataResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        console.error("error", error);
        reject(error);
      });
  });
};
export const createFlow = async (
  data: CreateFlowRequest,
): Promise<CreateFlowResponse> => {
  const newData = CreateFlowRequest.create({ ...data });

  return new Promise<CreateFlowResponse>((resolve, reject) => {
    makeGRPCCall<CreateFlowRequest, FlowServiceClient, CreateFlowResponse>(
      newData,
      flowServiceClient,
      "createFlow",
    )
      .then((response: CreateFlowResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const updateFlow = async (
  data: UpdateFlowRequest,
): Promise<UpdateFlowResponse> => {
  console.log("update flow data", data);
  return new Promise<UpdateFlowResponse>((resolve, reject) => {
    makeGRPCCall<UpdateFlowRequest, FlowServiceClient, UpdateFlowResponse>(
      data,
      flowServiceClient,
      "updateFlow",
    )
      .then((response: UpdateFlowResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readFlow = async (
  data: ReadFlowRequest,
): Promise<ReadFlowResponse> => {
  return new Promise<ReadFlowResponse>((resolve, reject) => {
    makeGRPCCall<ReadFlowRequest, FlowServiceClient, ReadFlowResponse>(
      data,
      flowServiceClient,
      "readFlow",
    )
      .then((response: ReadFlowResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// delete flow
export const deleteFlow = async (
  data: ReadFlowRequest,
): Promise<ReadFlowResponse> => {
  return new Promise<ReadFlowResponse>((resolve, reject) => {
    makeGRPCCall<DeleteFlowRequest, FlowServiceClient, DeleteFlowResponse>(
      data,
      flowServiceClient,
      "deleteFlow",
    )
      .then((response: ReadFlowResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readStage = async (
  data: ReadPipelineStageRequest,
): Promise<ReadPipelineStageResponse> => {
  return new Promise<ReadPipelineStageResponse>((resolve, reject) => {
    makeGRPCCall<
      ReadPipelineStageRequest,
      PipelineStageServiceClient,
      ReadPipelineStageResponse
    >(data, pipelineStageServiceClient, "readPipelineStage")
      .then((response: ReadPipelineStageResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// create stage labels

export const createStageLabel = async (
  data: CreateStageLabelRequest,
): Promise<CreateStageLabelResponse> => {
  console.log("label data", data);
  return new Promise<CreateStageLabelResponse>((resolve, reject) => {
    makeGRPCCall<
      CreateStageLabelRequest,
      StageLabelServiceClient,
      CreateStageLabelResponse
    >(data, stageLabelServiceClient, "createStageLabel")
      .then((response: CreateStageLabelResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// Read stage labels

export const readStageLabels = async (
  data: ReadStageLabelsRequest,
): Promise<ReadStageLabelsResponse> => {
  return new Promise<ReadStageLabelsResponse>((resolve, reject) => {
    makeGRPCCall<
      ReadStageLabelsRequest,
      StageLabelServiceClient,
      ReadStageLabelsResponse
    >(data, stageLabelServiceClient, "readStageLabels")
      .then((response: ReadStageLabelsResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// LEad CustomFields

export const createCustomField = async (
  data: CreateCustomFieldRequest,
): Promise<CreateCustomFieldResponse> => {
  return new Promise<CreateCustomFieldResponse>((resolve, reject) => {
    makeGRPCCall<
      CreateCustomFieldRequest,
      LeadServiceClient,
      CreateCustomFieldResponse
    >(data, leadServiceClient, "createCustomField")
      .then((response: CreateCustomFieldResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readCustomField = async (
  data: ReadCustomFieldRequest,
): Promise<ReadCustomFieldResponse> => {
  return new Promise<ReadCustomFieldResponse>((resolve, reject) => {
    makeGRPCCall<
      ReadCustomFieldRequest,
      LeadServiceClient,
      ReadCustomFieldResponse
    >(data, leadServiceClient, "readCustomField")
      .then((response: ReadCustomFieldResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readCustomFields = async (
  data: ReadLeadKeysRequest,
): Promise<ReadLeadKeysResponse> => {
  return new Promise<ReadLeadKeysResponse>((resolve, reject) => {
    makeGRPCCall<ReadLeadKeysRequest, LeadServiceClient, ReadLeadKeysResponse>(
      data,
      leadServiceClient,
      "readLeadKeys",
    )
      .then((response: ReadLeadKeysResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const updateCustomField = async (
  data: UpdateCustomFieldRequest,
): Promise<UpdateCustomFieldResponse> => {
  return new Promise<UpdateCustomFieldResponse>((resolve, reject) => {
    makeGRPCCall<
      UpdateCustomFieldRequest,
      LeadServiceClient,
      UpdateCustomFieldResponse
    >(data, leadServiceClient, "updateCustomField")
      .then((response: UpdateCustomFieldResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const deleteCustomField = async (
  data: DeleteCustomFieldRequest,
): Promise<DeleteCustomFieldResponse> => {
  return new Promise<DeleteCustomFieldResponse>((resolve, reject) => {
    makeGRPCCall<
      DeleteCustomFieldRequest,
      LeadServiceClient,
      DeleteCustomFieldResponse
    >(data, leadServiceClient, "deleteCustomField")
      .then((response: DeleteCustomFieldResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// Leads
export const createLead = async (
  data: CreateLeadRequest,
): Promise<CreateLeadResponse> => {
  return new Promise<CreateLeadResponse>((resolve, reject) => {
    makeGRPCCall<CreateLeadRequest, LeadServiceClient, CreateLeadResponse>(
      data,
      leadServiceClient,
      "createLead",
    )
      .then((response: CreateLeadResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readLead = async (
  data: ReadLeadRequest,
): Promise<ReadLeadResponse> => {
  return new Promise<ReadLeadResponse>((resolve, reject) => {
    makeGRPCCall<ReadLeadRequest, LeadServiceClient, ReadLeadResponse>(
      data,
      leadServiceClient,
      "readLead",
    )
      .then((response: ReadLeadResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const readLeads = async (
  data: ListLeadRequest,
): Promise<ListLeadResponse> => {
  return new Promise<ListLeadResponse>((resolve, reject) => {
    makeGRPCCall<ListLeadRequest, LeadServiceClient, ListLeadResponse>(
      data,
      leadServiceClient,
      "readLeads",
    )
      .then((response: ListLeadResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

export const bulkLeadsImport = async (
  data: BulkLeadsImportRequest,
): Promise<BulkLeadsImportResponse> => {
  return new Promise<BulkLeadsImportResponse>((resolve, reject) => {
    makeGRPCCall<
      BulkLeadsImportRequest,
      LeadServiceClient,
      BulkLeadsImportResponse
    >(data, leadServiceClient, "bulkLeadsImport")
      .then((response: BulkLeadsImportResponse) => {
        resolve(response);
      })
      .catch((error: RpcError) => {
        reject(error);
      });
  });
};

// TODO

// export const bulkLeadsExport = (
//   data: BulkLeadsExportRequest,
// ): Promise<Blob> => {
//   return new Promise<Blob>((resolve, reject) => {
//     const call = leadServiceClient.getService().bulkLeadsExport(data, {
//       meta: {
//         authorization: tokenObject().authorization,
//       },
//     });

//     const chunks: Uint8Array[] = [];

//     // Handle stream events
//     call.responses.onMessage((response: BulkLeadsExportResponse) => {
//       chunks.push(response.chunkData);
//     });

//     call.responses.onComplete(() => resolve(new Blob(chunks, { type: "text/csv" })));

//     call.responses.onError((reason: Error) => {
//       reject(reason);
//     });
//   });
// };
