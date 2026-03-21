import { useRef, useState } from 'react';
import { Box, Button, Typography, Chip } from '@mui/material';
import AttachFileIcon from '@mui/icons-material/AttachFile';

interface FileUploadProps {
  onUpload: (file: File) => Promise<void>;
}

const ACCEPTED_TYPES = '.xlsx,.xls,.pdf';
const MAX_SIZE = 10 * 1024 * 1024; // 10MB

export default function FileUpload({ onUpload }: FileUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState('');
  const [uploading, setUploading] = useState(false);

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > MAX_SIZE) {
      alert('ファイルサイズが大きすぎます（上限10MB）');
      return;
    }

    setUploading(true);
    setFileName(file.name);
    try {
      await onUpload(file);
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  return (
    <Box>
      <input
        ref={fileInputRef}
        type="file"
        accept={ACCEPTED_TYPES}
        hidden
        onChange={handleChange}
      />
      <Button
        variant="outlined"
        startIcon={<AttachFileIcon />}
        onClick={handleClick}
        disabled={uploading}
        size="small"
      >
        {uploading ? 'アップロード中...' : 'ファイル選択'}
      </Button>
      <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
        Excel(.xlsx/.xls) / PDF(.pdf) 最大10MB
      </Typography>
      {fileName && (
        <Box sx={{ mt: 1 }}>
          <Chip label={fileName} size="small" onDelete={() => setFileName('')} />
        </Box>
      )}
    </Box>
  );
}
