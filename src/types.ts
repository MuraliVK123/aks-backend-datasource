import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  target?: string;
  type?: string;
  checked?: boolean;
  alias?: string;
  scale?: string;
  pattern?: string;
  rawQuery?: boolean
}

export const defaultQuery: Partial<MyQuery> = {
 
  checked : false,
  type : "Log"

  //server: "ws://test:8080",
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  url?: string;
}

export interface MyVariableQuery {
  namespace: string;
  query: string;
  rawQuery?: string
}
