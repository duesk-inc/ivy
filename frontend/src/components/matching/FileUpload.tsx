import { useRef, useState, useCallback } from 'react';
import { Box, Typography, Chip } from '@mui/material';
import AttachFileIcon from '@mui/icons-material/AttachFile';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { ActionButton } from '../common';

interface FileUploadProps {
  onUpload: (file: File) => Promise<void>;
  fileName?: string;
  onClear?: () => void;
}

const ACCEPTED_TYPES = '.xlsx,.xls,.pdf';
const MAX_SIZE = 10 * 1024 * 1024; // 10MB

export default function FileUpload({ onUpload, fileName, onClear }: FileUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);

  const processFile = useCallback(async (file: File) => {
    if (file.size > MAX_SIZE) {
      alert('ファイルサイズが大きすぎます（上限10MB）');
      return;
    }

    const ext = file.name.toLowerCase().split('.').pop();
    if (!ext || !['xlsx', 'xls', 'pdf'].includes(ext)) {
      alert('対応していないファイル形式です（Excel / PDF のみ）');
      return;
    }

    setUploading(true);
    try {
      await onUpload(file);
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }, [onUpload]);

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) await processFile(file);
  };

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(false);
  }, []);

  const handleDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(false);

    const file = e.dataTransfer.files?.[0];
    if (file) await processFile(file);
  }, [processFile]);

  return (
    <Box>
      <input
        ref={fileInputRef}
        type="file"
        accept={ACCEPTED_TYPES}
        hidden
        onChange={handleChange}
      />

      {fileName ? (
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            gap: 1,
            p: 2,
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 2,
          }}
        >
          <AttachFileIcon fontSize="small" color="action" />
          <Chip
            label={fileName}
            size="small"
            onDelete={onClear}
          />
          <ActionButton
            buttonType="ghost"
            size="small"
            onClick={handleClick}
            loading={uploading}
          >
            変更
          </ActionButton>
        </Box>
      ) : (
        <Box
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={handleClick}
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 1,
            p: 3,
            border: '2px dashed',
            borderColor: dragOver ? 'primary.main' : 'divider',
            borderRadius: 2,
            bgcolor: dragOver ? 'action.hover' : 'transparent',
            cursor: 'pointer',
            transition: 'all 0.2s ease',
            '&:hover': {
              borderColor: 'primary.light',
              bgcolor: 'action.hover',
            },
          }}
        >
          {uploading ? (
            <Typography variant="body2" color="text.secondary">
              アップロード中...
            </Typography>
          ) : (
            <>
              <CloudUploadIcon sx={{ fontSize: 32, color: dragOver ? 'primary.main' : 'text.secondary' }} />
              <Typography variant="body2" color="text.secondary">
                ファイルをドラッグ&ドロップ、またはクリックして選択
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Excel(.xlsx/.xls) / PDF(.pdf) 最大10MB
              </Typography>
            </>
          )}
        </Box>
      )}
    </Box>
  );
}
