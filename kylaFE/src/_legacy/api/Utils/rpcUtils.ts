/* eslint-disable @typescript-eslint/no-explicit-any */
import type { RpcOptions } from "@protobuf-ts/runtime-rpc";
import { GRPCServiceClient } from "../globalClient/GlobalClients";

export const tokenObject = () => {
	const accessToken =
		JSON.parse(localStorage.getItem("recoil-persist") ?? "{}").authToken
			?.accessToken ?? "";
	return {
		authorization: accessToken,
		"content-type": "application/json",
	};
};

export const makeGRPCCall = async <T, C, S>(
	data: T,
	client: GRPCServiceClient<C>,
	method: string
): Promise<S> => {
	let res: Promise<S>;
	const options: RpcOptions = {
		meta: {
			authorization: tokenObject().authorization,
			"content-type": "application/grpc-web-text",
			"x-grpc-web": "1",
		},
	};
	try {
		res = (client.Service as C & any)[method](data, options).response;
		return res;
	} catch (error: any) {
		console.warn(`grpc call failed: ${error}`);
		throw error;
	}
};

export const streamGRPCCall = async <T, C, S>(
	data: T,
	client: GRPCServiceClient<C>,
	method: string,
	onData: (chunk: S) => void,
	onEnd?: () => void,
	onError?: (error: any) => void
): Promise<void> => {
	try {
		// let res: Promise<S>;
		const options: RpcOptions = {
			meta: {
				authorization: tokenObject().authorization,
				"x-grpc-web": "1",
				// "x-grpc-service": `${service ?? "core"}`,
			},
		};

		// console.log((client.Service as C & any)[method](data, options).on)
		const stream = (client.Service as C & any)[method](data, options);
		stream.on("data", (chunk: S) => {
			onData(chunk);
		});

		stream.on("end", () => {
			if (onEnd) onEnd();
		});

		stream.on("error", (error: any) => {
			if (onError) {
				onError(error);
			} else {
				console.error(`grpc stream eroor: ${error.message}`);
			}
		});
		// res = (client.Service as C & any)[method](data, options).response;
		// console.log(res)
		// return res;
	} catch (error: any) {
		console.error(`grpc call failed: ${error.message}`);

		if (onError) onError(error);
		throw error;
	}
};
