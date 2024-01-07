import React, {useEffect, useRef, useState} from 'react';
import {Button, Col, Form, Layout, Row} from "@douyinfe/semi-ui";
import VChart from '@visactor/vchart';
import {useEffectOnce} from "usehooks-ts";
import {API, isAdmin, showError, timestamp2string, timestamp2string1} from "../../helpers";
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
    const initialized = useRef(false)
    const [modelDataChart, setModelDataChart] = useState(null);
    const [modelDataPieChart, setModelDataPieChart] = useState(null);
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

    const loadQuotaData = async (lineChart, pieChart) => {
        setLoading(true);

        let url = '';
        let localStartTimestamp = Date.parse(start_timestamp) / 1000;
        let localEndTimestamp = Date.parse(end_timestamp) / 1000;
        if (isAdminUser) {
            url = `/api/data/?username=${username}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
        } else {
            url = `/api/data/self/?start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
        }
        const res = await API.get(url);
        const {success, message, data} = res.data;
        if (success) {
            setQuotaData(data);
            if (data.length === 0) {
                return;
            }
            updateChart(lineChart, pieChart, data);
        } else {
            showError(message);
        }
        setLoading(false);
    };

    const refresh = async () => {
        await loadQuotaData(modelDataChart, modelDataPieChart);
    };

    const initChart  = async () => {
        let lineChart = modelDataChart
        if (!modelDataChart) {
            lineChart = new VChart(spec_line, {dom: 'model_data'});
            setModelDataChart(lineChart);
            await lineChart.renderAsync();
        }
        let pieChart = modelDataPieChart
        if (!modelDataPieChart) {
            pieChart = new VChart(spec_pie, {dom: 'model_pie'});
            setModelDataPieChart(pieChart);
            await pieChart.renderAsync();
        }
        console.log('init vchart');
        await loadQuotaData(lineChart, pieChart)
    }

    const updateChart = (lineChart, pieChart, data) => {
        if (isAdminUser) {
            // 将所有用户合并
        }
        let pieData = [];
        let lineData = [];
        for (let i = 0; i < data.length; i++) {
            const item = data[i];
            // 合并model_name
            let pieItem = pieData.find(it => it.type === item.model_name);
            if (pieItem) {
                pieItem.value += item.count;
            } else {
                pieData.push({
                    "type": item.model_name,
                    "value": item.count
                });
            }
            // 合并created_at和model_name 为 lineData, created_at 数据类型是小时的时间戳
            // 转换日期格式
            let createTime = timestamp2string1(item.created_at);
            let lineItem = lineData.find(it => it.Time === createTime && it.Model === item.model_name);
            if (lineItem) {
                lineItem.Usage += parseFloat(getQuotaWithUnit(item.quota));
            } else {
                lineData.push({
                    "Time": createTime,
                    "Model": item.model_name,
                    "Usage": parseFloat(getQuotaWithUnit(item.quota))
                });
            }

        }
        // sort by count
        pieData.sort((a, b) => b.value - a.value);
        pieChart.updateData('id0', pieData);
        lineChart.updateData('barData', lineData);
        pieChart.reLayout();
        lineChart.reLayout();
    }

    useEffect(() => {
        if (!initialized.current) {
            initialized.current = true;
            initChart();
        }
    }, []);

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
                            {/*{*/}
                            {/*    isAdminUser && <>*/}
                            {/*        <Form.Input field="username" label='用户名称' style={{width: 176}} value={username}*/}
                            {/*                    placeholder={'可选值'} name='username'*/}
                            {/*                    onChange={value => handleInputChange(value, 'username')}/>*/}
                            {/*    </>*/}
                            {/*}*/}
                            <Form.Section>
                                <Button label='查询' type="primary" htmlType="submit" className="btn-margin-right"
                                        onClick={refresh}>查询</Button>
                            </Form.Section>
                        </>
                    </Form>
                    <div style={{height: 500}}>
                        <div id="model_pie" style={{width: '100%', minWidth: 100}}></div>
                    </div>
                    <div style={{height: 500}}>
                        <div id="model_data" style={{width: '100%', minWidth: 100}}></div>
                    </div>
                </Layout.Content>
            </Layout>
        </>
    );
};


export default Detail;
