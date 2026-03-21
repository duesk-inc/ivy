import React, { createContext, useContext, useState, useCallback } from "react";
import { Snackbar, Alert, Slide, useMediaQuery, useTheme } from "@mui/material";
import type { TransitionProps } from "@mui/material/transitions";

// PC: slide down from top
const SlideDownTransition = React.forwardRef<
  unknown,
  TransitionProps & { children: React.ReactElement }
>(function SlideDownTransition(props, ref) {
  return <Slide {...props} direction="down" ref={ref} />;
});

// Mobile: slide up from bottom
const SlideUpTransition = React.forwardRef<
  unknown,
  TransitionProps & { children: React.ReactElement }
>(function SlideUpTransition(props, ref) {
  return <Slide {...props} direction="up" ref={ref} />;
});

export interface ToastOptions {
  type: "success" | "error" | "warning" | "info";
  message: string;
  duration?: number;
  position?:
    | "top"
    | "bottom"
    | "top-right"
    | "top-center"
    | "bottom-right"
    | "bottom-center";
  title?: string;
  content?: React.ReactNode;
}

interface ToastState extends ToastOptions {
  id: string;
  open: boolean;
}

interface ToastContextValue {
  showToast: (options: ToastOptions) => void;
  showSuccess: (message: string, options?: Partial<ToastOptions>) => void;
  showError: (message: string, options?: Partial<ToastOptions>) => void;
  showWarning: (message: string, options?: Partial<ToastOptions>) => void;
  showInfo: (message: string, options?: Partial<ToastOptions>) => void;
  hideToast: () => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

export interface ToastProviderProps {
  children: React.ReactNode;
  defaultDuration?: number;
  defaultPosition?:
    | "top"
    | "bottom"
    | "top-right"
    | "top-center"
    | "bottom-right"
    | "bottom-center";
}

export const ToastProvider: React.FC<ToastProviderProps> = ({
  children,
  defaultDuration = 6000,
  defaultPosition = "top-center",
}) => {
  const [toast, setToast] = useState<ToastState | null>(null);

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("sm"));

  const showToast = useCallback(
    (options: ToastOptions) => {
      const id = Math.random().toString(36).substr(2, 9);
      setToast({
        id,
        open: true,
        duration: defaultDuration,
        position: defaultPosition,
        ...options,
      });
    },
    [defaultDuration, defaultPosition],
  );

  const showSuccess = useCallback(
    (message: string, options?: Partial<ToastOptions>) => {
      if (!message) {
        console.warn("[Toast] Success message is empty or undefined");
        return;
      }
      showToast({ type: "success", message, ...options });
    },
    [showToast],
  );

  const showError = useCallback(
    (message: string, options?: Partial<ToastOptions>) => {
      if (!message) {
        console.warn("[Toast] Error message is empty or undefined");
        return;
      }
      showToast({ type: "error", message, ...options });
    },
    [showToast],
  );

  const showWarning = useCallback(
    (message: string, options?: Partial<ToastOptions>) => {
      if (!message) {
        console.warn("[Toast] Warning message is empty or undefined");
        return;
      }
      showToast({ type: "warning", message, ...options });
    },
    [showToast],
  );

  const showInfo = useCallback(
    (message: string, options?: Partial<ToastOptions>) => {
      if (!message) {
        console.warn("[Toast] Info message is empty or undefined");
        return;
      }
      showToast({ type: "info", message, ...options });
    },
    [showToast],
  );

  const hideToast = useCallback(() => {
    setToast((prev) => (prev ? { ...prev, open: false } : null));
  }, []);

  const handleClose = useCallback(
    (_event?: React.SyntheticEvent | Event, reason?: string) => {
      if (reason === "clickaway") {
        return;
      }
      hideToast();
    },
    [hideToast],
  );

  const value: ToastContextValue = {
    showToast,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    hideToast,
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
      {toast && (
        <Snackbar
          open={toast.open}
          autoHideDuration={toast.duration}
          onClose={handleClose}
          TransitionComponent={
            isMobile ? SlideUpTransition : SlideDownTransition
          }
          sx={
            isMobile
              ? { bottom: "var(--toast-bottom-offset, 0px) !important" }
              : undefined
          }
          anchorOrigin={(() => {
            if (isMobile) {
              return { vertical: "bottom", horizontal: "center" as const };
            }
            const pos = toast.position || defaultPosition || "top-right";
            const vertical =
              pos.startsWith("top") || pos === "top" ? "top" : "bottom";
            let horizontal: "left" | "center" | "right" = "right";
            if (pos.endsWith("center") || pos === "top" || pos === "bottom")
              horizontal = "center";
            if (pos.endsWith("right")) horizontal = "right";
            return { vertical, horizontal };
          })()}
        >
          <Alert
            onClose={handleClose}
            severity={toast.type}
            sx={{ width: "100%", whiteSpace: "pre-line" }}
            data-testid={`toast-${toast.type}`}
          >
            {toast.title && (
              <div style={{ fontWeight: "bold", marginBottom: 4 }}>
                {toast.title}
              </div>
            )}
            {toast.content ?? toast.message}
          </Alert>
        </Snackbar>
      )}
    </ToastContext.Provider>
  );
};

export const useToast = (): ToastContextValue => {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
};
