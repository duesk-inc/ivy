import { useState, useRef } from 'react';
import {
  Box,
  Grid,
  TextField,
  Typography,
  Card,
  CardContent,
  Alert,
  Divider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Tooltip,
} from '@mui/material';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import LinkIcon from '@mui/icons-material/Link';
import { ActionButton } from '../components/common';
import { SimpleTextField, SimpleSelect } from '../components/common/forms';
import Layout from '../components/common/Layout';
import MatchingResult from '../components/matching/MatchingResult';
import FileUpload from '../components/matching/FileUpload';
import { useToast } from '../components/common/Toast';
import { executeMatching, parseFile, createJobGroup } from '../lib/api/client';
import type { MatchingRequest, MatchingResponse, SupplementInfo } from '../types';

export default function MatchingPage() {
  const [jobText, setJobText] = useState('');
  const [engineerText, setEngineerText] = useState('');
  const [engineerFileText, setEngineerFileText] = useState('');
  const [engineerFileKey, setEngineerFileKey] = useState('');
  const [engineerFileName, setEngineerFileName] = useState('');
  const [supplement, setSupplement] = useState<SupplementInfo>({});
  const [result, setResult] = useState<MatchingResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const resultRef = useRef<HTMLDivElement>(null);
  const { showSuccess, showError: showErrorToast } = useToast();

  const handleFileUpload = async (file: File) => {
    try {
      const response = await parseFile(file);
      setEngineerFileKey(response.file_key);
      setEngineerFileName(file.name);
      setEngineerFileText(response.text);
      if (response.parse_warnings?.length > 0) {
        setError(response.parse_warnings.join('\n'));
      }
    } catch (err: any) {
      setError(err?.response?.data?.error || 'ファイルの読み取りに失敗しました');
    }
  };

  const handleFileClear = () => {
    setEngineerFileKey('');
    setEngineerFileName('');
    setEngineerFileText('');
  };

  const handleExecute = async () => {
    if (!jobText.trim()) {
      setError('案件情報を入力してください');
      return;
    }
    if (!engineerText.trim() && !engineerFileText.trim()) {
      setError('エンジニア情報（テキストまたはファイル）を入力してください');
      return;
    }

    setError('');
    setLoading(true);
    setResult(null);

    try {
      // ファイル抽出テキストと補足テキストを結合
      const combinedEngineerText = [engineerFileText, engineerText]
        .filter((t) => t.trim())
        .join('\n\n--- 補足情報 ---\n');

      const req: MatchingRequest = {
        job_text: jobText,
        engineer_text: combinedEngineerText,
        engineer_file_key: engineerFileKey || undefined,
        supplement: Object.keys(supplement).length > 0 ? supplement : undefined,
      };
      const response = await executeMatching(req);
      setResult(response);
      setTimeout(() => {
        resultRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 100);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'マッチング処理に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const updateSupplement = (key: keyof SupplementInfo, value: any) => {
    setSupplement((prev) => {
      if (value === '' || value === undefined) {
        const next = { ...prev };
        delete next[key];
        return next;
      }
      return { ...prev, [key]: value };
    });
  };

  return (
    <Layout>
      <Typography variant="h5" sx={{ mb: 3, fontWeight: 600 }}>
        個別マッチング
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* 案件情報 - multiline textarea: keep MUI TextField */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
                案件情報
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                メール本文をそのまま貼り付けてください
              </Typography>
              <TextField
                multiline
                rows={10}
                fullWidth
                placeholder={"【案件】Java開発\n【単価】〜70万\n【場所】東京都渋谷区\n【時期】即日\n【必須スキル】Java 3年以上\n\n※メール本文をそのまま貼り付けてOKです"}
                value={jobText}
                onChange={(e) => setJobText(e.target.value)}
              />
            </CardContent>
          </Card>
        </Grid>

        {/* エンジニア情報 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
                エンジニア情報
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
                スキルシート（Excel/PDF）をアップロード
              </Typography>
              <FileUpload
                onUpload={handleFileUpload}
                fileName={engineerFileName}
                onClear={handleFileClear}
              />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 2, mb: 1 }}>
                スキルシートに記載のない追加情報（任意）
              </Typography>
              <TextField
                multiline
                rows={5}
                fullWidth
                placeholder={"例: 希望単価60万、即日稼働可、AWS実務経験あり（経歴書未記載）\n\n※配信メールの本文を貼り付けてもOKです"}
                value={engineerText}
                onChange={(e) => setEngineerText(e.target.value)}
              />
            </CardContent>
          </Card>
        </Grid>

        {/* 補足情報 */}
        <Grid size={12}>
          <Accordion defaultExpanded variant="outlined" sx={{ borderRadius: 2, '&:before': { display: 'none' } }}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                補足情報
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Grid container spacing={2}>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleSelect
                    value={supplement.affiliation_type || ''}
                    onChange={(v) => updateSupplement('affiliation_type', v)}
                    options={[
                      { value: '', label: '未選択' },
                      { value: 'duesk', label: 'デュスク社員' },
                      { value: 'partner', label: 'パートナー' },
                    ]}
                    label="所属"
                    size="small"
                  />
                </Grid>
                {supplement.affiliation_type === 'partner' && (
                  <Grid size={{ xs: 6, sm: 2 }}>
                    <SimpleTextField
                      label="パートナー会社名"
                      size="small"
                      fullWidth
                      value={supplement.affiliation_name || ''}
                      onChange={(v) => updateSupplement('affiliation_name', v)}
                    />
                  </Grid>
                )}
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleTextField
                    label="希望単価（万円）"
                    type="number"
                    size="small"
                    fullWidth
                    value={supplement.rate || ''}
                    onChange={(v) => updateSupplement('rate', v ? Number(v) : undefined)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleSelect
                    value={supplement.nationality || ''}
                    onChange={(v) => updateSupplement('nationality', v)}
                    options={[
                      { value: '', label: '未選択' },
                      { value: 'japanese', label: '日本' },
                      { value: 'other', label: '外国籍' },
                    ]}
                    label="国籍"
                    size="small"
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleSelect
                    value={supplement.employment_type || ''}
                    onChange={(v) => updateSupplement('employment_type', v)}
                    options={[
                      { value: '', label: '未選択' },
                      { value: 'employee', label: '正社員' },
                      { value: 'freelance', label: 'フリーランス' },
                    ]}
                    label="雇用形態"
                    size="small"
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleTextField
                    label="稼働可能日"
                    placeholder="2026-04"
                    size="small"
                    fullWidth
                    value={supplement.available_from || ''}
                    onChange={(v) => updateSupplement('available_from', v)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <SimpleTextField
                    label="送信元会社名"
                    size="small"
                    fullWidth
                    value={supplement.supply_chain_source || ''}
                    onChange={(v) => updateSupplement('supply_chain_source', v)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <SimpleSelect<string | number>
                    value={supplement.supply_chain_level ?? ''}
                    onChange={(v) => {
                      updateSupplement('supply_chain_level', v === '' ? undefined : Number(v));
                    }}
                    options={[
                      { value: '', label: '未選択' },
                      { value: 1, label: 'エンド直' },
                      { value: 2, label: '1次請け' },
                      { value: 3, label: '2次請け' },
                      { value: 4, label: '3次以上' },
                    ]}
                    label="商流"
                    size="small"
                  />
                </Grid>
              </Grid>
            </AccordionDetails>
          </Accordion>
        </Grid>

        {/* 実行ボタン */}
        <Grid size={12}>
          <Box sx={{ display: 'flex', justifyContent: 'center' }}>
            <ActionButton
              buttonType="primary"
              size="large"
              icon={<CompareArrowsIcon />}
              onClick={handleExecute}
              loading={loading}
              sx={{ px: 6, py: 1.5 }}
            >
              {loading ? 'AIが分析中です（最大1分程度かかります）' : 'マッチング実行'}
            </ActionButton>
          </Box>
        </Grid>

        {/* 結果表示 */}
        {result && (
          <Grid size={12} ref={resultRef}>
            <Divider sx={{ my: 2 }} />
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
              <Tooltip title="同じ案件に複数エンジニアを比較する際に使用します" arrow>
                <span>
                  <ActionButton
                    buttonType="secondary"
                    size="small"
                    icon={<LinkIcon />}
                    onClick={async () => {
                      const name = window.prompt('案件グループ名を入力してください');
                      if (name) {
                        try {
                          await createJobGroup(name, result.id);
                          showSuccess('案件グループを作成しました');
                        } catch {
                          showErrorToast('作成に失敗しました');
                        }
                      }
                    }}
                  >
                    案件グループを作成
                  </ActionButton>
                </span>
              </Tooltip>
            </Box>
            <MatchingResult result={result} />
          </Grid>
        )}
      </Grid>
    </Layout>
  );
}
