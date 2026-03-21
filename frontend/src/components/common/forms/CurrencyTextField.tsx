import React, { useCallback } from 'react';
import { TextField, InputAdornment } from '@mui/material';
import type { SxProps, Theme } from '@mui/material';

interface CurrencyTextFieldProps {
  value: number | string;
  onChange: (value: number) => void;
  label?: string;
  currencyPosition?: 'start' | 'end';
  min?: number;
  max?: number;
  step?: number;
  disabled?: boolean;
  required?: boolean;
  placeholder?: string;
  size?: 'small' | 'medium';
  fullWidth?: boolean;
  error?: boolean;
  helperText?: string;
  sx?: SxProps<Theme>;
  name?: string;
}

function CurrencyTextField({
  value,
  onChange,
  label,
  currencyPosition = 'start',
  min = 0,
  max,
  step = 1,
  disabled = false,
  required = false,
  placeholder,
  size = 'medium',
  fullWidth = true,
  error = false,
  helperText,
  sx,
  name,
}: CurrencyTextFieldProps) {
  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const inputValue = event.target.value;
      if (inputValue === '' || inputValue === '-') {
        onChange(0);
        return;
      }
      const numericValue = parseFloat(inputValue);
      if (isNaN(numericValue)) {
        onChange(0);
        return;
      }
      onChange(numericValue);
    },
    [onChange]
  );

  const buildInputProps = () => {
    if (currencyPosition === 'start') {
      return {
        startAdornment: <InputAdornment position="start">¥</InputAdornment>,
      };
    }
    return {
      endAdornment: <InputAdornment position="end">円</InputAdornment>,
    };
  };

  const displayValue = typeof value === 'number' && value === 0 ? '' : value;

  return (
    <TextField
      value={displayValue}
      onChange={handleChange}
      name={name}
      label={label}
      type="number"
      disabled={disabled}
      required={required}
      placeholder={placeholder ?? '0'}
      size={size}
      fullWidth={fullWidth}
      error={error}
      helperText={helperText}
      InputProps={buildInputProps()}
      inputProps={{ min, step, ...(max !== undefined && { max }) }}
      sx={{
        '& input[type=number]': {
          MozAppearance: 'textfield',
        },
        '& input[type=number]::-webkit-outer-spin-button, & input[type=number]::-webkit-inner-spin-button':
          {
            WebkitAppearance: 'none',
            margin: 0,
          },
        ...sx,
      }}
    />
  );
}

export default CurrencyTextField;
