import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { SectionLoader } from './components/common';
import { AuthProvider, useAuth } from './context/AuthContext';

const LoginPage = lazy(() => import('./pages/LoginPage'));
const MatchingPage = lazy(() => import('./pages/MatchingPage'));
const HistoryPage = lazy(() => import('./pages/HistoryPage'));
const MatchingDetailPage = lazy(() => import('./pages/MatchingDetailPage'));
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const JobsPage = lazy(() => import('./pages/JobsPage'));
const EngineersPage = lazy(() => import('./pages/EngineersPage'));
const BatchMatchingPage = lazy(() => import('./pages/BatchMatchingPage'));

function Loading() {
  return <SectionLoader padding={8} />;
}

interface ProtectedRouteProps {
  children: React.ReactNode;
  adminOnly?: boolean;
}

function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();

  if (isLoading) {
    return <Loading />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (adminOnly && user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MatchingPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/history"
          element={
            <ProtectedRoute>
              <HistoryPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/history/:id"
          element={
            <ProtectedRoute>
              <MatchingDetailPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/jobs"
          element={
            <ProtectedRoute>
              <JobsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/engineers"
          element={
            <ProtectedRoute>
              <EngineersPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/batch-matching"
          element={
            <ProtectedRoute>
              <BatchMatchingPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute adminOnly>
              <SettingsPage />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}
