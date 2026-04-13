export type ApiRequest = <T>(path: string, options?: RequestInit) => Promise<T>;
