import { Box, Typography, LinearProgress } from '@mui/material';

interface ScoreBarProps {
  label: string;
  score?: number;
  max?: number;
  reason?: string;
}

export default function ScoreBar({ label, score = 0, max = 10, reason }: ScoreBarProps) {
  const percentage = max > 0 ? (score / max) * 100 : 0;

  const getColor = () => {
    if (percentage >= 80) return 'success';
    if (percentage >= 50) return 'primary';
    if (percentage >= 30) return 'warning';
    return 'error';
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
        <Typography variant="body2" sx={{ fontWeight: 600 }}>
          {label}
        </Typography>
        <Typography variant="body2" sx={{ fontWeight: 600 }}>
          {score} / {max}
        </Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={percentage}
        color={getColor() as any}
        sx={{ height: 8, borderRadius: 4, mb: 0.5 }}
      />
      {reason && (
        <Typography variant="caption" color="text.secondary">
          {reason}
        </Typography>
      )}
    </Box>
  );
}
