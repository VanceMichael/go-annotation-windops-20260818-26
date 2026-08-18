export type ApiError={code:string;message:string;request_id:string};
export class ClientError extends Error{constructor(public status:number,public detail:ApiError){super(detail.message)}}
const tenant='demo';
export async function request<T>(path:string,init:RequestInit={},signal?:AbortSignal):Promise<T>{const response=await fetch(path,{...init,signal,headers:{'Content-Type':'application/json','X-Tenant-ID':tenant,...init.headers}});const body=await response.json().catch(()=>({}));if(!response.ok)throw new ClientError(response.status,body.error??{code:'invalid_response',message:'服务返回了无法识别的错误',request_id:response.headers.get('X-Request-ID')??''});return body as T}
export type EntityPage<T>={items:T[];total:number};
export const api={overview:(signal?:AbortSignal)=>request<Record<string,number>>('/api/overview',{},signal),rules:(signal?:AbortSignal)=>request<{items:string[]}>('/api/rules',{},signal),entities:<T>(name:string,signal?:AbortSignal)=>request<EntityPage<T>>('/api/'+name,{},signal),decide:<T>(payload:T,signal?:AbortSignal)=>request('/api/decisions',{method:'POST',body:JSON.stringify(payload)},signal)};

