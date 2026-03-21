import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Chip,
  Alert,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import type { MatchingResponse, MatchResult } from '../../types';
import { gradeColor, skillStatusColor, skillStatusLabel, ngStatusColor, ngStatusLabel } from '../../utils/grade';
import ScoreBar from './ScoreBar';

interface MatchingResultProps {
  result: MatchingResponse | (MatchingResponse & { job_text?: string; engineer_text?: string });
}

const NG_FLAG_LABELS: Record<string, string> = {
  nationality: '外国籍',
  freelancer: '個人事業主',
  supply_chain: '商流',
  age: '年齢',
};

export default function MatchingResult({ result }: MatchingResultProps) {
  const r = result.result as MatchResult;
  if (!r) return null;

  return (
    <Box>
      {/* ヘッダー: 総合スコア */}
      <Card sx={{ mb: 3, bgcolor: gradeColor(result.grade), color: 'white' }}>
        <CardContent sx={{ textAlign: 'center', py: 3 }}>
          <Typography variant="h3" sx={{ fontWeight: 700 }}>
            {result.total_score} / 100
          </Typography>
          <Typography variant="h5" sx={{ mt: 1 }}>
            {result.grade}判定 - {result.grade_label}
          </Typography>
        </CardContent>
      </Card>

      <Grid container spacing={3}>
        {/* 案件サマリー・エンジニアサマリー */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                案件情報
              </Typography>
              <Table size="small">
                <TableBody>
                  <SummaryRow label="案件名" value={r.job_summary?.name} />
                  <SummaryRow label="勤務地" value={r.job_summary?.location} />
                  <SummaryRow label="リモート" value={r.job_summary?.remote} />
                  <SummaryRow label="単価" value={r.job_summary?.rate} />
                  <SummaryRow label="開始" value={r.job_summary?.start} />
                  <SummaryRow label="条件" value={r.job_summary?.conditions} />
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                エンジニア情報
              </Typography>
              <Table size="small">
                <TableBody>
                  <SummaryRow label="イニシャル" value={r.engineer_summary?.initials} />
                  <SummaryRow label="年齢" value={r.engineer_summary?.age ? `${r.engineer_summary.age}歳` : undefined} />
                  <SummaryRow label="性別" value={r.engineer_summary?.gender} />
                  <SummaryRow label="最寄駅" value={r.engineer_summary?.nearest_station} />
                  <SummaryRow label="所属" value={r.engineer_summary?.affiliation} />
                  <SummaryRow label="希望単価" value={r.engineer_summary?.rate} />
                  <SummaryRow label="稼働可能日" value={r.engineer_summary?.available_from} />
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>

        {/* スコア詳細 */}
        <Grid size={12}>
          <Card>
            <CardContent>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 2 }}>
                スコア詳細
              </Typography>
              {r.scores && (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                  <ScoreBar label="スキル適合" score={r.scores.skill?.score} max={r.scores.skill?.max} reason={r.scores.skill?.reason} />
                  <ScoreBar label="稼働時期" score={r.scores.timing?.score} max={r.scores.timing?.max} reason={r.scores.timing?.reason} />
                  <ScoreBar label="単価" score={r.scores.rate?.score} max={r.scores.rate?.max} reason={r.scores.rate?.reason} />
                  <ScoreBar label="経験年数" score={r.scores.experience_years?.score} max={r.scores.experience_years?.max} reason={r.scores.experience_years?.reason} />
                  <ScoreBar label="勤務地" score={r.scores.location?.score} max={r.scores.location?.max} reason={r.scores.location?.reason} />
                  <ScoreBar label="業界経験" score={r.scores.industry?.score} max={r.scores.industry?.max} reason={r.scores.industry?.reason} />
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* スキル詳細（折りたたみ） */}
        {r.scores?.skill?.required_skills && (
          <Grid size={12}>
            <Accordion variant="outlined" sx={{ borderRadius: 2, '&:before': { display: 'none' } }}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                  スキル判定（詳細）
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                  必須スキル
                </Typography>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
                  {r.scores.skill.required_skills.map((s, i) => (
                    <Chip
                      key={i}
                      label={`${s.skill}: ${skillStatusLabel(s.status)}`}
                      size="small"
                      sx={{ bgcolor: skillStatusColor(s.status), color: 'white' }}
                      title={s.detail}
                    />
                  ))}
                </Box>
                {r.scores.skill.optional_skills && r.scores.skill.optional_skills.length > 0 && (
                  <>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                      尚可スキル
                    </Typography>
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                      {r.scores.skill.optional_skills.map((s, i) => (
                        <Chip
                          key={i}
                          label={`${s.skill}: ${skillStatusLabel(s.status)}`}
                          size="small"
                          variant="outlined"
                          title={s.detail}
                        />
                      ))}
                    </Box>
                  </>
                )}
              </AccordionDetails>
            </Accordion>
          </Grid>
        )}

        {/* NG判定 */}
        {r.ng_flags && (
          <Grid size={{ xs: 12, md: 6 }}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 2 }}>
                  NG判定
                </Typography>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                  {Object.entries(r.ng_flags).map(([key, flag]) => (
                    <Chip
                      key={key}
                      label={`${NG_FLAG_LABELS[key] || key}: ${ngStatusLabel(flag.status)}`}
                      size="small"
                      sx={{ bgcolor: ngStatusColor(flag.status), color: 'white' }}
                      title={flag.detail}
                    />
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        )}

        {/* ポジティブ・ネガティブ */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              {r.positives && r.positives.length > 0 && (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="subtitle2" color="success.main" sx={{ fontWeight: 600, mb: 1 }}>
                    ポジティブ要素
                  </Typography>
                  {r.positives.map((p, i) => (
                    <Typography key={i} variant="body2" sx={{ mb: 0.5 }}>
                      + {p}
                    </Typography>
                  ))}
                </Box>
              )}
              {r.negatives && r.negatives.length > 0 && (
                <Box>
                  <Typography variant="subtitle2" color="error.main" sx={{ fontWeight: 600, mb: 1 }}>
                    懸念事項
                  </Typography>
                  {r.negatives.map((n, i) => (
                    <Typography key={i} variant="body2" sx={{ mb: 0.5 }}>
                      - {n}
                    </Typography>
                  ))}
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* 警告 */}
        {r.warnings && r.warnings.length > 0 && (
          <Grid size={12}>
            {r.warnings.map((w, i) => (
              <Alert key={i} severity="warning" sx={{ mb: 1 }}>
                {w}
              </Alert>
            ))}
          </Grid>
        )}

        {/* アドバイス */}
        {r.advice && (
          <Grid size={12}>
            <Card>
              <CardContent>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                  アドバイス
                </Typography>
                <Typography variant="body2">{r.advice}</Typography>
              </CardContent>
            </Card>
          </Grid>
        )}

        {/* 確認ヒント */}
        {r.confirmation_hints && r.confirmation_hints.length > 0 && (
          <Grid size={12}>
            <Card>
              <CardContent>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 2 }}>
                  追加確認のヒント
                </Typography>
                {r.confirmation_hints.map((hint, i) => (
                  <Paper key={i} variant="outlined" sx={{ p: 2, mb: i < r.confirmation_hints.length - 1 ? 2 : 0 }}>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                      対象: {hint.target}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1, fontWeight: 500 }}>
                      {hint.question}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      根拠: {hint.reason}
                    </Typography>
                  </Paper>
                ))}
              </CardContent>
            </Card>
          </Grid>
        )}
      </Grid>
    </Box>
  );
}

function SummaryRow({ label, value }: { label: string; value?: string | number }) {
  if (!value) return null;
  return (
    <TableRow>
      <TableCell sx={{ border: 0, py: 0.5, pl: 0, width: 100 }}>
        <Typography variant="body2" color="text.secondary">{label}</Typography>
      </TableCell>
      <TableCell sx={{ border: 0, py: 0.5 }}>
        <Typography variant="body2">{value}</Typography>
      </TableCell>
    </TableRow>
  );
}
