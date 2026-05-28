import { RpcError } from "grpc-web";
import { makeGRPCCall } from "./rpcUtils";
import type { CreateBillingAccountRequest, CreateBillingAccountResponse, ReadBillingAccountRequest, ReadBillingAccountResponse } from "@/pb/billing_accounts";
import type { BillingAccountsServiceClient } from "@/pb/billing_accounts.client";
import type { ReadPaymentMethodsRequest, ReadPaymentMethodsResponse } from "@/pb/billing_payment_methods";
import type { PaymentMethodsServiceClient } from "@/pb/billing_payment_methods.client";
import type { SubscribeToPlanRequest, SubscribeToPlanResponse, CancelSubscriptionRequest, CancelSubscriptionResponse } from "@/pb/billing_subscription";
import type { SubscriptionServiceClient } from "@/pb/billing_subscription.client";
import type { ReadSubscriptionPlansRequest, ReadSubscriptionPlansResponse } from "@/pb/billing_subscription_plans";
import type { SubscriptionPlansServiceClient } from "@/pb/billing_subscription_plans.client";
import type { TopUpWalletRequest, TopUpWalletResponse, CreateWalletRequest, CreateWalletResponse, ReadWalletRequest, ReadWalletResponse, ReadWalletTransactionsRequest, ReadWalletTransactionsResponse } from "@/pb/billing_wallets";
import type { WalletsServiceClient } from "@/pb/billing_wallets.client";
import { subscriptionPlansServiceClient, subscriptionServiceClient, walletsServiceClient, billingAccountsServiceClient, paymentMethodsServiceClient } from "../globalClient/GlobalClients";



export const readSubscriptionPlans = (
  data: ReadSubscriptionPlansRequest
): Promise<ReadSubscriptionPlansResponse> => {
  return new Promise<ReadSubscriptionPlansResponse>((resolve, reject) => {
    makeGRPCCall<ReadSubscriptionPlansRequest, SubscriptionPlansServiceClient, ReadSubscriptionPlansResponse>(
      data,
      subscriptionPlansServiceClient,
      "readSubscriptionPlans"
    )
      .then((response: ReadSubscriptionPlansResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const subscribeToPlan = (
  data: SubscribeToPlanRequest
): Promise<SubscribeToPlanResponse> => {
  return new Promise<SubscribeToPlanResponse>((resolve, reject) => {
    makeGRPCCall<SubscribeToPlanRequest, SubscriptionServiceClient, SubscribeToPlanResponse>(
      data,
      subscriptionServiceClient,
      "subscribeToPlan"
    )
      .then((response: SubscribeToPlanResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}


export const cancelSubscription = (
  data: CancelSubscriptionRequest
): Promise<CancelSubscriptionResponse> => {
  return new Promise<CancelSubscriptionResponse>((resolve, reject) => {
    makeGRPCCall<CancelSubscriptionRequest, SubscriptionServiceClient, CancelSubscriptionResponse>(
      data,
      subscriptionServiceClient,
      "cancelSubscription"
    )
      .then((response: CancelSubscriptionResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const topUpWallet = (
  data: TopUpWalletRequest
): Promise<TopUpWalletResponse> => {
  return new Promise<TopUpWalletResponse>((resolve, reject) => {
    makeGRPCCall<TopUpWalletRequest, WalletsServiceClient, TopUpWalletResponse>(
      data,
      walletsServiceClient,
      "topUpWallet"
    )
      .then((response: TopUpWalletResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const createWallet = (
  data: CreateWalletRequest
): Promise<CreateWalletResponse> => {
  return new Promise<CreateWalletResponse>((resolve, reject) => {
    makeGRPCCall<CreateWalletRequest, WalletsServiceClient, CreateWalletResponse>(
      data,
      walletsServiceClient,
      "createWallet"
    )
      .then((response: CreateWalletResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const readWallet = (
  data: ReadWalletRequest
): Promise<ReadWalletResponse> => {
  return new Promise<ReadWalletResponse>((resolve, reject) => {
    makeGRPCCall<ReadWalletRequest, WalletsServiceClient, ReadWalletResponse>(
      data,
      walletsServiceClient,
      "readWallet"
    )
      .then((response: ReadWalletResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      })
  })
}

export const readWalletTransactions = (
  data: ReadWalletTransactionsRequest
): Promise<ReadWalletTransactionsResponse> => {
  return new Promise<ReadWalletTransactionsResponse>((resolve, reject) => {
    makeGRPCCall<ReadWalletTransactionsRequest, WalletsServiceClient, ReadWalletTransactionsResponse>(
      data,
      walletsServiceClient,
      "readWalletTransactions"
    )
      .then((response: ReadWalletTransactionsResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      })
  })
}

export const createBillingAccount = (
  data: CreateBillingAccountRequest
): Promise<CreateBillingAccountResponse> => {
  return new Promise<CreateBillingAccountResponse>((resolve, reject) => {
    makeGRPCCall<CreateBillingAccountRequest, BillingAccountsServiceClient, CreateBillingAccountResponse>(
      data,
      billingAccountsServiceClient,
      "createBillingAccount"
    )
      .then((response: CreateBillingAccountResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const readBillingAccount = (
  data: ReadBillingAccountRequest
): Promise<ReadBillingAccountResponse> => {
  return new Promise<ReadBillingAccountResponse>((resolve, reject) => {
    makeGRPCCall<ReadBillingAccountRequest, BillingAccountsServiceClient, ReadBillingAccountResponse>(
      data,
      billingAccountsServiceClient,
      "readBillingAccount"
    )
      .then((response: ReadBillingAccountResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

export const readAllPaymentMethods = (
  data: ReadPaymentMethodsRequest
): Promise<ReadPaymentMethodsResponse> => {
  return new Promise<ReadPaymentMethodsResponse>((resolve, reject) => {
    makeGRPCCall<ReadPaymentMethodsRequest, PaymentMethodsServiceClient, ReadPaymentMethodsResponse>(
      data,
      paymentMethodsServiceClient,
      "readAllPaymentMethods"
    )
      .then((response: ReadPaymentMethodsResponse) => {
        resolve(response)
      })
      .catch((error: RpcError) => {
        reject(error)
      });
  });
}

