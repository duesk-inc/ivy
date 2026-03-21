import React from 'react';
import { TextField, InputAdornment } from '@mui/material';
import type { SxProps, Theme } from '@mui/material';

interface SimpleTextFieldProps {
  value: string | number;
  onChange: (value: string) => void;
  label: string;
  type?: React.HTMLInputTypeAttribute;
  disabled?: boolean;
  required?: boolean;
  placeholder?: string;
  size?: 'small' | 'medium';
  fullWidth?: boolean;
  multiline?: boolean;
  rows?: number;
  maxLength?: number;
  startAdornment?: React.ReactNode;
  endAdornment?: React.ReactNode;
  error?: boolean;
  helperText?: string;
  autoFocus?: boolean;
  autoComplete?: string;
  readOnly?: boolean;
  sx?: SxProps<Theme>;
  name?: string;
  onBlur?: () => void;
}

function SimpleTextField({
  value,
  onChange,
  label,
  type = 'text',
  disabled = false,
  required = false,
  placeholder,
  size = 'medium',
  fullWidth = true,
  multiline = false,
  rows,
  maxLength,
  startAdornment,
  endAdornment,
  error = false,
  helperText,
  autoFocus = false,
  autoComplete,
  readOnly = false,
  sx,
  name,
  onBlur,
}: SimpleTextFieldProps) {
  const buildInputProps = () => {
    const inputProps: Record<string, unknown> = {};
    if (startAdornment) {
      inputProps.startAdornment = (
        <InputAdornment position="start">{startAdornment}</InputAdornment>
      );
    }
    if (endAdornment) {
      inputProps.endAdornment = (
        <InputAdornment position="end">{endAdornment}</InputAdornment>
      );
    }
    if (readOnly) {
      inputProps.readOnly = true;
    }
    return inputProps;
  };

  const buildInputElementProps = () => {
    const props: Record<string, unknown> = {};
    if (maxLength) {
      props.maxLength = maxLength;
    }
    return props;
  };

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(event.target.value);
  };

  return (
    <TextField
      value={value}
      onChange={handleChange}
      onBlur={onBlur}
      name={name}
      label={label}
      type={type}
      disabled={disabled}
      required={required}
      placeholder={placeholder}
      size={size}
      fullWidth={fullWidth}
      multiline={multiline}
      rows={multiline ? rows : undefined}
      error={error}
      helperText={helperText}
      autoFocus={autoFocus}
      autoComplete={autoComplete}
      InputProps={buildInputProps()}
      inputProps={buildInputElementProps()}
      InputLabelProps={{
        shrink: ['date', 'time', 'datetime-local'].includes(type) ? true : undefined,
      }}
      sx={sx}
    />
  );
}

export default SimpleTextField;
