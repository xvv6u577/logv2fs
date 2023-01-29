// types.d.ts:
export interface  TrafficByPeriod {
	amount: number;
	period: number;
	used_by_domain: { [key: string]: number };
}

interface NodeInUseStatus {
	[domain: string]: boolean;
}

export interface User {
	name: string;
	email: string;
	role: string;
	used:number;
	path: string;
	uuid: string;
	status: string;
	password: string;
	credit: string;
	used: string;
	node_in_use_status: NodeInUseStatus[];
	used_by_current_month: TrafficByPeriod;
	used_by_current_day: TrafficByPeriod;
	traffic_by_day: TrafficByPeriod[];
	traffic_by_month: TrafficByPeriod[];
} 

export interface LoginState {
	isLogin: boolean;
	jwt: any;
	token: string;
}

export interface MessageState {
	show: boolean;
	type: string;
	content: string;
}

export interface MessagePayload {
	show: boolean;
	content: string;
}

export interface RerenderState {
	rerender: boolean;
}

export interface IStore {
	login: LoginState;
	message: MessageState;
	rerender: RerenderState;
}