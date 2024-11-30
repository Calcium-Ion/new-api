import { Input, Typography } from '@douyinfe/semi-ui';
import React from 'react';

const TextInput = ({ label, name, value, onChange, placeholder, type = 'text' }) => {
  return (
    <>
      <div style={{ marginTop: 10 }}>
        <Typography.Text strong>{label}</Typography.Text>
      </div>
      <Input
        name={name}
        placeholder={placeholder}
        onChange={(value) => onChange(value)}
        value={value}
        autoComplete="new-password"
      />
    </>
  );
}

export default TextInput;