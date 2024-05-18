import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';

export default function SettingsMagnification(props) {
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ModelPrice: '',
    ModelRatio: '',
    CompletionRatio: '',
    GroupRatio: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  async function onSubmit() {
    try {
      await refForm.current.validate();
      const updateArray = compareObjects(inputs, inputsRow);
      if (!updateArray.length) return showWarning('你似乎并没有修改什么');
      const requestQueue = updateArray.map((item) => {
        let value = '';
        if (typeof inputs[item.key] === 'boolean') {
          value = String(inputs[item.key]);
        } else {
          value = inputs[item.key];
        }
        return API.put('/api/option/', {
          key: item.key,
          value,
        });
      });
      setLoading(true);
      Promise.all(requestQueue)
        .then((res) => {
          if (requestQueue.length === 1) {
            if (res.includes(undefined)) return;
          } else if (requestQueue.length > 1) {
            if (res.includes(undefined))
              return showError('部分保存失败，请重试');
          }
          showSuccess('保存成功');
          props.refresh();
        })
        .catch(() => {
          showError('保存失败，请重试');
        })
        .finally(() => {
          setLoading(false);
        });
    } catch (error) {
      showError('请检查输入');
      console.error(error);
    } finally {
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);
  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={'倍率设置'}>
            <Row gutter={16}>
              <Col span={16}>
                <Form.TextArea
                  label={'模型固定价格'}
                  extraText={'一次调用消耗多少刀，优先级大于模型倍率'}
                  placeholder={
                    '为一个 JSON 文本，键为模型名称，值为一次调用消耗多少刀，比如 "gpt-4-gizmo-*": 0.1，一次消耗0.1刀'
                  }
                  field={'ModelPrice'}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: '不是合法的 JSON 字符串',
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      ModelPrice: value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={16}>
                <Form.TextArea
                  label={'模型倍率'}
                  extraText={''}
                  placeholder={'为一个 JSON 文本，键为模型名称，值为倍率'}
                  field={'ModelRatio'}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: '不是合法的 JSON 字符串',
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      ModelRatio: value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={16}>
                <Form.TextArea
                  label={'模型补全倍率（仅对自定义模型有效）'}
                  extraText={'仅对自定义模型有效'}
                  placeholder={'为一个 JSON 文本，键为模型名称，值为倍率'}
                  field={'CompletionRatio'}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: '不是合法的 JSON 字符串',
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      CompletionRatio: value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={16}>
                <Form.TextArea
                  label={'分组倍率'}
                  extraText={''}
                  placeholder={'为一个 JSON 文本，键为分组名称，值为倍率'}
                  field={'GroupRatio'}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: '不是合法的 JSON 字符串',
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      GroupRatio: value,
                    })
                  }
                />
              </Col>
            </Row>

            <Row>
              <Button size='large' onClick={onSubmit}>
                保存倍率设置
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
