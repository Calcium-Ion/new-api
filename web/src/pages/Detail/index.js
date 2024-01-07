import React, {useEffect, useState} from 'react';
import {Button, Col, Form, Layout, Row} from "@douyinfe/semi-ui";
import VChart from '@visactor/vchart';
import {useEffectOnce} from "usehooks-ts";
import {API, isAdmin, showError, timestamp2string, timestamp2string1} from "../../helpers";
import {ITEMS_PER_PAGE} from "../../constants";
import {getQuotaWithUnit} from "../../helpers/render";

const Detail = (props) => {

    let now = new Date();
    const [inputs, setInputs] = useState({
        username: '',
        token_name: '',
        model_name: '',
        start_timestamp: timestamp2string(now.getTime() / 1000 - 86400),
        end_timestamp: timestamp2string(now.getTime() / 1000 + 3600),
        channel: ''
    });
    const {username, token_name, model_name, start_timestamp, end_timestamp, channel} = inputs;
    const isAdminUser = isAdmin();
    let modelDataChart = null;
    let modelDataPieChart = null;
    const [loading, setLoading] = useState(true);
    const [quotaData, setQuotaData] = useState([]);
    const [quotaDataPie, setQuotaDataPie] = useState([]);
    const [quotaDataLine, setQuotaDataLine] = useState([]);

    const handleInputChange = (value, name) => {
        setInputs((inputs) => ({...inputs, [name]: value}));
    };

    const spec_line = {
        type: 'bar',
        data: [
            {
                id: 'barData',
                values: [
                ]
            }
        ],
        xField: 'Time',
        yField: 'Usage',
        seriesField: 'Model',
        stack: true,
        legends: {
            visible: true
        },
        title: {
            visible: true,
            text: '模型消耗分布'
        },
        bar: {
            // The state style of bar
            state: {
                hover: {
                    stroke: '#000',
                    lineWidth: 1
                }
            }
        }
    };

    const spec_pie = {
        type: 'pie',
        data: [
            {
                id: 'id0',
                values: [
                    { type: 'null', value: '0' },
                ]
            }
        ],
        outerRadius: 0.8,
        innerRadius: 0.5,
        padAngle: 0.6,
        valueField: 'value',
        categoryField: 'type',
        pie: {
            style: {
                cornerRadius: 10
            },
            state: {
                hover: {
                    outerRadius: 0.85,
                    stroke: '#000',
                    lineWidth: 1
                },
                selected: {
                    outerRadius: 0.85,
                    stroke: '#000',
                    lineWidth: 1
                }
            }
        },
        title: {
            visible: true,
            text: '模型调用次数占比'
        },
        legends: {
            visible: true,
            orient: 'left'
        },
        label: {
            visible: true
        },
        tooltip: {
            mark: {
                content: [
                    {
                        key: datum => datum['type'],
                        value: datum => datum['value']
                    }
                ]
            }
        }
    };

    const loadQuotaData = async () => {
        setLoading(true);

        let url = '';
        let localStartTimestamp = Date.parse(start_timestamp) / 1000;
        let localEndTimestamp = Date.parse(end_timestamp) / 1000;
        if (isAdminUser) {
            url = `/api/data`;
        } else {
            url = `/api/data/self`;
        }
        const res = await API.get(url);
        const {success, message, data} = res.data;
        if (success) {
            setQuotaData(data);
            updateChart(data);
        } else {
            showError(message);
        }
        setLoading(false);
    };

    const refresh = async () => {
        await loadQuotaData();
    };

    const updateChart = (data) => {
        if (isAdminUser) {
            // 将所有用户的数据累加
            let pieData = [];
            let lineData = [];
            for (let i = 0; i < data.length; i++) {
                const item = data[i];
                const {count, id, model_name, quota, user_id, username} = item;
                // 合并model_name
                let pieItem = pieData.find(item => item.model_name === model_name);
                if (pieItem) {
                    pieItem.count += count;
                } else {
                    pieData.push({
                        "type": model_name,
                        "value": count
                    });
                }
                // 合并created_at和model_name 为 lineData, created_at 数据类型是小时的时间戳
                // 转换日期格式
                let createTime = timestamp2string1(item.created_at);
                let lineItem = lineData.find(item => item.Time === item.createTime && item.Model === model_name);
                if (lineItem) {
                    lineItem.Usage += getQuotaWithUnit(quota);
                } else {
                    lineData.push({
                        "Time": createTime,
                        "Model": model_name,
                        "Usage": getQuotaWithUnit(quota)
                    });
                }

            }
            // sort by count
            pieData.sort((a, b) => b.value - a.value);
            spec_line.data[0].values = lineData;
            spec_pie.data[0].values = pieData;
            // console.log('spec_line', spec_line);
            console.log('spec_pie', spec_pie);
            // modelDataChart.renderAsync();
            modelDataPieChart.updateSpec(spec_pie);
            modelDataChart.updateSpec(spec_line);
        }
    }

    useEffect(() => {
        refresh();
    }, []);

    useEffectOnce(() => {
        // 创建 vchart 实例
        if (!modelDataChart) {
            modelDataChart = new VChart(spec_line, {dom: 'model_data'});
            // 绘制
            modelDataChart.renderAsync();
        }

        if (!modelDataPieChart) {
            modelDataPieChart = new VChart(spec_pie, {dom: 'model_pie'});
            // 绘制
            modelDataPieChart.renderAsync();
        }

        console.log('render vchart');
    })

    return (
        <>
            <Layout>
                <Layout.Header>
                    <h3>数据看板(24H)</h3>
                </Layout.Header>
                <Layout.Content>
                    <Form layout='horizontal' style={{marginTop: 10}}>
                        <>
                            <Form.DatePicker field="start_timestamp" label='起始时间' style={{width: 272}}
                                             initValue={start_timestamp}
                                             value={start_timestamp} type='dateTime'
                                             name='start_timestamp'
                                             onChange={value => handleInputChange(value, 'start_timestamp')}/>
                            <Form.DatePicker field="end_timestamp" fluid label='结束时间' style={{width: 272}}
                                             initValue={end_timestamp}
                                             value={end_timestamp} type='dateTime'
                                             name='end_timestamp'
                                             onChange={value => handleInputChange(value, 'end_timestamp')}/>
                            {/*<Form.Button fluid label='操作' width={2} onClick={refresh}>查询</Form.Button>*/}
                            {/*{*/}
                            {/*    isAdminUser && <>*/}
                            {/*        <Form.Input field="username" label='用户名称' style={{width: 176}} value={username}*/}
                            {/*                    placeholder={'可选值'} name='username'*/}
                            {/*                    onChange={value => handleInputChange(value, 'username')}/>*/}
                            {/*    </>*/}
                            {/*}*/}
                            {/*<Form.Section>*/}
                            {/*    <Button label='查询' type="primary" htmlType="submit" className="btn-margin-right"*/}
                            {/*            >查询</Button>*/}
                            {/*</Form.Section>*/}
                        </>
                    </Form>
                    <div style={{height: 500}}>
                        <div id="model_pie" style={{width: '100%'}}></div>
                    </div>
                    <div style={{height: 500}}>
                        <div id="model_data" style={{width: '100%'}}></div>
                    </div>
                </Layout.Content>
            </Layout>
        </>
    );
};


export default Detail;
