import { Input, InputNumber, Typography } from '@douyinfe/semi-ui';
import React from 'react';

const TextNumberInput = ({ label, name, value, onChange, placeholder }) => {
  return (
    <>
      <div style={{ marginTop: 10 }}>
        <Typography.Text strong>{label}</Typography.Text>
      </div>
      <InputNumber
        name={name}
        placeholder={placeholder}
        onChange={(value) => onChange(value)}
        value={value}
        autoComplete="new-password"
      />
    </>
  );
}

export default TextNumberInput;