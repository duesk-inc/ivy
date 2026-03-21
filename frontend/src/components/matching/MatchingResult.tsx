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
} from '@mui/material';
import type { MatchingResponse, MatchResult } from '../../types';
import { gradeColor, skillStatusColor, ngStatusColor, ngStatusLabel } from '../../utils/grade';
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

  // NG/warning のみ抽出
  const ngIssues = r.ng_flags
    ? Object.entries(r.ng_flags).filter(([, f]) => f.status === 'ng' || f.status === 'warning')
    : [];

  return (
    <Box>
      {/* ヘッダー: 総合スコア */}
      <Card sx={{ mb: 2, bgcolor: gradeColor(result.grade), color: 'white' }}>
        <CardContent sx={{ textAlign: 'center', py: 2 }}>
          <Typography variant="h3" sx={{ fontWeight: 700 }}>
            {result.total_score} / 100
          </Typography>
          <Typography variant="h6" sx={{ mt: 0.5 }}>
            {result.grade}判定 - {result.grade_label}
          </Typography>
        </CardContent>
      </Card>

      <Grid container spacing={2}>

        {/* === 案件情報 + エンジニア情報（スコア直下） === */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.5 }}>案件情報</Typography>
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
            <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.5 }}>エンジニア情報</Typography>
              <Table size="small">
                <TableBody>
                  <SummaryRow label="イニシャル" value={r.engineer_summary?.initials} />
                  <SummaryRow label="年齢" value={r.engineer_summary?.age ? `${r.engineer_summary.age}歳` : undefined} />
                  <SummaryRow label="最寄駅" value={r.engineer_summary?.nearest_station} />
                  <SummaryRow label="所属" value={r.engineer_summary?.affiliation} />
                  <SummaryRow label="希望単価" value={r.engineer_summary?.rate} />
                  <SummaryRow label="稼働可能日" value={r.engineer_summary?.available_from} />
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>

        {/* === スキル要件 + NG/ポジティブ/ネガティブ === */}
        <Grid size={{ xs: 12, md: 7 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                スキル要件
              </Typography>
              {r.scores?.skill?.required_skills && (
                <>
                  <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                    必須スキル
                  </Typography>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75, mb: 1.5 }}>
                    {r.scores.skill.required_skills.map((s, i) => (
                      <Chip
                        key={i}
                        label={s.skill}
                        size="small"
                        sx={{
                          bgcolor: s.status === 'met' ? skillStatusColor('met')
                            : s.status === 'partial' ? skillStatusColor('partial')
                            : '#e0e0e0',
                          color: s.status === 'unmet' ? 'text.secondary' : 'white',
                          fontWeight: 500,
                        }}
                        title={s.detail}
                      />
                    ))}
                  </Box>
                </>
              )}
              {r.scores?.skill?.optional_skills && r.scores.skill.optional_skills.length > 0 && (
                <>
                  <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                    尚可スキル
                  </Typography>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75 }}>
                    {r.scores.skill.optional_skills.map((s, i) => (
                      <Chip
                        key={i}
                        label={s.skill}
                        size="small"
                        sx={{
                          bgcolor: s.status === 'met' ? skillStatusColor('met')
                            : s.status === 'partial' ? skillStatusColor('partial')
                            : '#e0e0e0',
                          color: s.status === 'unmet' ? 'text.secondary' : 'white',
                          fontWeight: 500,
                        }}
                        title={s.detail}
                      />
                    ))}
                  </Box>
                </>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 5 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, height: '100%' }}>
            {/* NG判定: NGやwarningがある場合のみ表示 */}
            {ngIssues.length > 0 ? (
              <Card sx={{ border: '2px solid', borderColor: 'error.main' }}>
                <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1, color: 'error.main' }}>
                    NG該当あり
                  </Typography>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75 }}>
                    {ngIssues.map(([key, flag]) => (
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
            ) : (
              <Card>
                <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 600, color: 'success.main' }}>
                    NG該当なし
                  </Typography>
                </CardContent>
              </Card>
            )}

            {/* ポジティブ・ネガティブ */}
            <Card sx={{ flex: 1 }}>
              <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                {r.positives && r.positives.length > 0 && (
                  <Box sx={{ mb: r.negatives?.length ? 1.5 : 0 }}>
                    <Typography variant="subtitle2" color="success.main" sx={{ fontWeight: 600, mb: 0.5 }}>
                      ポジティブ要素
                    </Typography>
                    {r.positives.map((p, i) => (
                      <Typography key={i} variant="body2" sx={{ mb: 0.25 }}>+ {p}</Typography>
                    ))}
                  </Box>
                )}
                {r.negatives && r.negatives.length > 0 && (
                  <Box>
                    <Typography variant="subtitle2" color="error.main" sx={{ fontWeight: 600, mb: 0.5 }}>
                      懸念事項
                    </Typography>
                    {r.negatives.map((n, i) => (
                      <Typography key={i} variant="body2" sx={{ mb: 0.25 }}>- {n}</Typography>
                    ))}
                  </Box>
                )}
              </CardContent>
            </Card>
          </Box>
        </Grid>

        {/* === 警告 === */}
        {r.warnings && r.warnings.length > 0 && (
          <Grid size={12}>
            {r.warnings.map((w, i) => (
              <Alert key={i} severity="warning" sx={{ mb: 1 }}>{w}</Alert>
            ))}
          </Grid>
        )}

        {/* === スコア詳細（2カラム） === */}
        {r.scores && (
          <Grid size={12}>
            <Card>
              <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1.5 }}>
                  スコア詳細
                </Typography>
                <Grid container spacing={2}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                      <ScoreBar label="スキル適合" score={r.scores.skill?.score} max={r.scores.skill?.max} reason={r.scores.skill?.reason} />
                      <ScoreBar label="稼働時期" score={r.scores.timing?.score} max={r.scores.timing?.max} reason={r.scores.timing?.reason} />
                      <ScoreBar label="単価" score={r.scores.rate?.score} max={r.scores.rate?.max} reason={r.scores.rate?.reason} />
                    </Box>
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                      <ScoreBar label="経験年数" score={r.scores.experience_years?.score} max={r.scores.experience_years?.max} reason={r.scores.experience_years?.reason} />
                      <ScoreBar label="勤務地" score={r.scores.location?.score} max={r.scores.location?.max} reason={r.scores.location?.reason} />
                      <ScoreBar label="業界経験" score={r.scores.industry?.score} max={r.scores.industry?.max} reason={r.scores.industry?.reason} />
                    </Box>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>
        )}

        {/* === アドバイス === */}
        {r.advice && (
          <Grid size={12}>
            <Card>
              <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 0.5 }}>
                  アドバイス
                </Typography>
                <Typography variant="body2">{r.advice}</Typography>
              </CardContent>
            </Card>
          </Grid>
        )}

        {/* === 確認ヒント === */}
        {r.confirmation_hints && r.confirmation_hints.length > 0 && (
          <Grid size={12}>
            <Card>
              <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1.5 }}>
                  追加確認のヒント
                </Typography>
                {r.confirmation_hints.map((hint, i) => (
                  <Paper key={i} variant="outlined" sx={{ p: 1.5, mb: i < r.confirmation_hints.length - 1 ? 1.5 : 0 }}>
                    <Typography variant="caption" color="text.secondary">
                      対象: {hint.target}
                    </Typography>
                    <Typography variant="body2" sx={{ fontWeight: 500, my: 0.5 }}>
                      {hint.question}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
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
      <TableCell sx={{ border: 0, py: 0.25, pl: 0, width: 80 }}>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
      </TableCell>
      <TableCell sx={{ border: 0, py: 0.25 }}>
        <Typography variant="body2">{value}</Typography>
      </TableCell>
    </TableRow>
  );
}
