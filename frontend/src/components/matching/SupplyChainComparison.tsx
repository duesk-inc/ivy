import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Chip,
} from '@mui/material';
import StarIcon from '@mui/icons-material/Star';
import type { JobGroup, MatchingListItem } from '../../types';
import { SUPPLY_CHAIN_LABELS } from '../../types';
import { gradeColor, supplyChainColor } from '../../utils/grade';

interface SupplyChainComparisonProps {
  jobGroup: JobGroup;
}

function MatchingCard({ item, isBestRoute }: { item: MatchingListItem; isBestRoute: boolean }) {
  return (
    <Card
      sx={{
        height: '100%',
        border: isBestRoute ? '2px solid #2E7D32' : '1px solid',
        borderColor: isBestRoute ? '#2E7D32' : 'divider',
      }}
    >
      <CardContent>
        {isBestRoute && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
            <StarIcon sx={{ color: '#2E7D32', fontSize: 20 }} />
            <Typography variant="body2" sx={{ color: '#2E7D32', fontWeight: 600 }}>
              推奨ルート
            </Typography>
          </Box>
        )}
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">送信元</Typography>
            <Typography variant="body2">{item.supply_chain_source || '-'}</Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">商流</Typography>
            <Chip
              label={SUPPLY_CHAIN_LABELS[item.supply_chain_level] || '不明'}
              size="small"
              sx={{
                bgcolor: supplyChainColor(item.supply_chain_level),
                color: 'white',
                fontWeight: 600,
              }}
            />
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">案件単価</Typography>
            <Typography variant="body2">{item.job_summary?.rate || '-'}</Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">スコア</Typography>
            <Typography variant="body2" fontWeight={600}>{item.total_score}点</Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">判定</Typography>
            <Chip
              label={`${item.grade} - ${item.grade_label}`}
              size="small"
              sx={{
                bgcolor: gradeColor(item.grade),
                color: 'white',
                fontWeight: 600,
              }}
            />
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
}

export default function SupplyChainComparison({ jobGroup }: SupplyChainComparisonProps) {
  return (
    <Box>
      <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
        {jobGroup.name}
      </Typography>

      {jobGroup.best_route && (
        <Box sx={{ mb: 3 }}>
          <Typography variant="subtitle2" sx={{ mb: 1, color: '#2E7D32', fontWeight: 600 }}>
            推奨ルート
          </Typography>
          <Box sx={{ maxWidth: 400 }}>
            <MatchingCard item={jobGroup.best_route} isBestRoute />
          </Box>
        </Box>
      )}

      <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>
        全ルート比較
      </Typography>
      <Grid container spacing={2}>
        {jobGroup.matchings.map((matching) => (
          <Grid key={matching.id} size={{ xs: 12, sm: 6, md: 4 }}>
            <MatchingCard
              item={matching}
              isBestRoute={jobGroup.best_route?.id === matching.id}
            />
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
