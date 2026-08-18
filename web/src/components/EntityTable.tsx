import {Table,Tag,Typography} from 'antd';import type {ColumnsType} from 'antd/es/table';
type Row={id:string;status:string;updated_at?:string;[key:string]:unknown};
const colors:Record<string,string>={active:'green',approved:'green',completed:'green',available:'cyan',draft:'default',planned:'blue',submitted:'gold',in_progress:'processing',failed:'red',canceled:'default'};
export function EntityTable({rows,details}:{rows:Row[];details:string[]}){const columns:ColumnsType<Row>=[{title:'编号',dataIndex:'id',width:180,render:value=><Typography.Text copyable>{value}</Typography.Text>},{title:'状态',dataIndex:'status',width:130,render:value=><Tag color={colors[value]??'default'}>{value}</Tag>},...details.map(key=>({title:key.replaceAll('_',' '),dataIndex:key,ellipsis:true})),{title:'更新时间',dataIndex:'updated_at',width:190}];return <Table rowKey="id" size="middle" columns={columns} dataSource={rows} pagination={{pageSize:10,showSizeChanger:false}} scroll={{x:900}}/>}

