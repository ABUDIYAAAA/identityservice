"use client";

import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";

export default function StateInspectModal({ open, onClose, event }) {
  if (!event) return null;

  const hasBefore = Boolean(
    event.before_state && Object.keys(event.before_state).length > 0
  );
  const hasAfter = Boolean(
    event.after_state && Object.keys(event.after_state).length > 0
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      slotProps={{
        paper: {
          elevation: 0,
          sx: {
            borderRadius: 3,
            p: 1,
            border: "1px solid",
            borderColor: "divider",
          },
        },
      }}
    >
      <DialogTitle fontWeight={700}>
        Audit Event State Inspection ({event.action_type})
      </DialogTitle>

      <DialogContent sx={{ pt: 1 }}>
        <Box sx={{ mb: 3 }}>
          <Typography variant="caption" color="text.secondary" display="block">
            Event ID: {event.id || "-"} | Timestamp:{" "}
            {event.created_at
              ? new Date(event.created_at).toLocaleString()
              : "-"}
          </Typography>
          {event.error_message && (
            <Typography
              variant="body2"
              color="error.main"
              sx={{ mt: 0.5, fontWeight: 600 }}
            >
              Error: {event.error_message}
            </Typography>
          )}
        </Box>

        <Box sx={{ display: "flex", gap: 2, flexDirection: { xs: "column", md: "row" } }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography
              variant="subtitle2"
              fontWeight={600}
              gutterBottom
              color="text.secondary"
            >
              Before State (Previous)
            </Typography>
            <Paper
              elevation={0}
              sx={{
                p: 2,
                backgroundColor: "#f8fafc",
                border: "1px solid",
                borderColor: "divider",
                borderRadius: 2,
                fontFamily: "monospace",
                fontSize: "0.8rem",
                maxHeight: 300,
                overflow: "auto",
              }}
            >
              {hasBefore ? (
                <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
                  {JSON.stringify(event.before_state, null, 2)}
                </pre>
              ) : (
                <Typography variant="caption" color="text.secondary">
                  No previous state recorded
                </Typography>
              )}
            </Paper>
          </Box>

          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography
              variant="subtitle2"
              fontWeight={600}
              gutterBottom
              color="text.secondary"
            >
              After State (Updated)
            </Typography>
            <Paper
              elevation={0}
              sx={{
                p: 2,
                backgroundColor: "#f8fafc",
                border: "1px solid",
                borderColor: "divider",
                borderRadius: 2,
                fontFamily: "monospace",
                fontSize: "0.8rem",
                maxHeight: 300,
                overflow: "auto",
              }}
            >
              {hasAfter ? (
                <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
                  {JSON.stringify(event.after_state, null, 2)}
                </pre>
              ) : (
                <Typography variant="caption" color="text.secondary">
                  No updated state recorded
                </Typography>
              )}
            </Paper>
          </Box>
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} variant="outlined" color="inherit">
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
}
