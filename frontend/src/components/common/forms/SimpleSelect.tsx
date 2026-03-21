import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormHelperText,
} from '@mui/material';
import type { SelectChangeEvent, SxProps, Theme } from '@mui/material';
import type { SelectOption } from '../../../types/forms';

interface SimpleSelectProps<T extends string | number = string> {
  value: T | '';
  onChange: (value: T) => void;
  options: SelectOption<T>[];
  label?: string;
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

function SimpleSelect<T extends string | number = string>({
  value,
  onChange,
  options,
  label,
  disabled = false,
  required = false,
  placeholder,
  size = 'medium',
  fullWidth = true,
  error = false,
  helperText,
  sx,
  name,
}: SimpleSelectProps<T>) {
  const labelId = label
    ? `${name || label.replace(/\s/g, '-').toLowerCase()}-label`
    : undefined;

  const handleChange = (event: SelectChangeEvent<T | ''>) => {
    onChange(event.target.value as T);
  };

  return (
    <FormControl
      fullWidth={fullWidth}
      error={error}
      size={size}
      disabled={disabled}
      sx={sx}
    >
      {label && (
        <InputLabel id={labelId} required={required}>
          {label}
        </InputLabel>
      )}
      <Select
        labelId={labelId}
        value={value}
        onChange={handleChange}
        label={label}
        name={name}
        MenuProps={{
          PaperProps: {
            sx: { maxHeight: 280, mt: 0.5 },
          },
        }}
      >
        {placeholder && (
          <MenuItem value="" disabled>
            {placeholder}
          </MenuItem>
        )}
        {options.map((option) => (
          <MenuItem
            key={String(option.value)}
            value={option.value}
            disabled={option.disabled}
          >
            {option.label}
          </MenuItem>
        ))}
      </Select>
      {helperText && <FormHelperText>{helperText}</FormHelperText>}
    </FormControl>
  );
}

export default SimpleSelect;
