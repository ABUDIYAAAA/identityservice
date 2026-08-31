"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import Box from "@mui/material/Box";
import Drawer from "@mui/material/Drawer";
import Typography from "@mui/material/Typography";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import Divider from "@mui/material/Divider";
import Avatar from "@mui/material/Avatar";
import Chip from "@mui/material/Chip";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import Button from "@mui/material/Button";
import { useAuth } from "@/hooks/useAuth";
import UserSettingsModal from "./UserSettingsModal";

const DRAWER_WIDTH = 260;

export default function DashboardLayout({ children }) {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  const [anchorEl, setAnchorEl] = useState(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const openUserMenu = Boolean(anchorEl);

  const handleOpenUserMenu = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleCloseUserMenu = () => {
    setAnchorEl(null);
  };

  const handleSignOut = () => {
    handleCloseUserMenu();
    logout();
  };

  const userInitial = user?.email ? user.email.charAt(0).toUpperCase() : "U";

  return (
    <Box
      sx={{
        display: "flex",
        minHeight: "100vh",
        backgroundColor: "background.default",
      }}
    >
      {/* Sidebar Drawer */}
      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          "& .MuiDrawer-paper": {
            width: DRAWER_WIDTH,
            boxSizing: "border-box",
            borderRight: "1px solid",
            borderColor: "divider",
            backgroundColor: "#ffffff",
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          },
        }}
      >
        {/* Top Section: Brand & Navigation */}
        <Box>
          <Box sx={{ p: 3, display: "flex", alignItems: "center", gap: 1.5 }}>
            <Box
              sx={{
                width: 32,
                height: 32,
                borderRadius: 2,
                backgroundColor: "primary.main",
                color: "#ffffff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontWeight: 700,
                fontSize: "1.1rem",
              }}
            >
              I
            </Box>
            <Typography variant="h6" fontWeight={700} color="text.primary">
              Identity
            </Typography>
          </Box>

          <Divider sx={{ mb: 2 }} />

          <List disablePadding sx={{ px: 1.5 }}>
            <ListItem disablePadding>
              <ListItemButton
                component={Link}
                href="/"
                selected={pathname === "/"}
                sx={{
                  borderRadius: 2,
                  mb: 0.5,
                  "&.Mui-selected": {
                    backgroundColor: "#f1f5f9",
                    color: "text.primary",
                    fontWeight: 600,
                    "&:hover": {
                      backgroundColor: "#e2e8f0",
                    },
                  },
                }}
              >
                <ListItemText
                  primary={
                    <Typography
                      sx={{
                        fontSize: "0.9rem",
                        fontWeight: pathname === "/" ? 600 : 500,
                      }}
                    >
                      Dashboard
                    </Typography>
                  }
                />
              </ListItemButton>
            </ListItem>

            {user?.role === "admin" && (
              <>
                <ListItem disablePadding>
                  <ListItemButton
                    component={Link}
                    href="/users"
                    selected={pathname.startsWith("/users")}
                    sx={{
                      borderRadius: 2,
                      mb: 0.5,
                      "&.Mui-selected": {
                        backgroundColor: "#f1f5f9",
                        color: "text.primary",
                        fontWeight: 600,
                        "&:hover": {
                          backgroundColor: "#e2e8f0",
                        },
                      },
                    }}
                  >
                    <ListItemText
                      primary={
                        <Typography
                          sx={{
                            fontSize: "0.9rem",
                            fontWeight: pathname.startsWith("/users")
                              ? 600
                              : 500,
                          }}
                        >
                          Users
                        </Typography>
                      }
                    />
                  </ListItemButton>
                </ListItem>

                <ListItem disablePadding>
                  <ListItemButton
                    component={Link}
                    href="/audit-logs"
                    selected={pathname.startsWith("/audit-logs")}
                    sx={{
                      borderRadius: 2,
                      mb: 0.5,
                      "&.Mui-selected": {
                        backgroundColor: "#f1f5f9",
                        color: "text.primary",
                        fontWeight: 600,
                        "&:hover": {
                          backgroundColor: "#e2e8f0",
                        },
                      },
                    }}
                  >
                    <ListItemText
                      primary={
                        <Typography
                          sx={{
                            fontSize: "0.9rem",
                            fontWeight: pathname.startsWith("/audit-logs")
                              ? 600
                              : 500,
                          }}
                        >
                          Audit Logs
                        </Typography>
                      }
                    />
                  </ListItemButton>
                </ListItem>
              </>
            )}

            <ListItem disablePadding>
              <ListItemButton
                component={Link}
                href="/services"
                selected={pathname.startsWith("/services")}
                sx={{
                  borderRadius: 2,
                  mb: 0.5,
                  "&.Mui-selected": {
                    backgroundColor: "#f1f5f9",
                    color: "text.primary",
                    fontWeight: 600,
                    "&:hover": {
                      backgroundColor: "#e2e8f0",
                    },
                  },
                }}
              >
                <ListItemText
                  primary={
                    <Typography
                      sx={{
                        fontSize: "0.9rem",
                        fontWeight: pathname.startsWith("/services")
                          ? 600
                          : 500,
                      }}
                    >
                      Services
                    </Typography>
                  }
                />
              </ListItemButton>
            </ListItem>
          </List>
        </Box>

        {/* Bottom Section: User Details Card & Sign Out Menu */}
        <Box sx={{ p: 2, borderTop: "1px solid", borderColor: "divider" }}>
          <Box
            onClick={handleOpenUserMenu}
            sx={{
              p: 1.5,
              borderRadius: 2.5,
              display: "flex",
              alignItems: "center",
              gap: 1.5,
              cursor: "pointer",
              transition: "background-color 0.2s",
              "&:hover": {
                backgroundColor: "#f8fafc",
              },
            }}
          >
            <Avatar
              sx={{
                width: 38,
                height: 38,
                backgroundColor: "primary.main",
                fontSize: "0.95rem",
                fontWeight: 600,
              }}
            >
              {userInitial}
            </Avatar>

            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography
                variant="body2"
                fontWeight={600}
                noWrap
                color="text.primary"
              >
                {user?.email}
              </Typography>
              <Chip
                label={user?.role || "user"}
                size="small"
                sx={{
                  height: 18,
                  fontSize: "0.65rem",
                  fontWeight: 600,
                  textTransform: "uppercase",
                  backgroundColor: "#f1f5f9",
                  color: "#475569",
                  mt: 0.2,
                }}
              />
            </Box>
          </Box>

          {/* User Popover / Menu */}
          <Menu
            anchorEl={anchorEl}
            open={openUserMenu}
            onClose={handleCloseUserMenu}
            anchorOrigin={{
              vertical: "top",
              horizontal: "center",
            }}
            transformOrigin={{
              vertical: "bottom",
              horizontal: "center",
            }}
            slotProps={{
              paper: {
                elevation: 0,
                sx: {
                  width: 230,
                  borderRadius: 3,
                  border: "1px solid",
                  borderColor: "divider",
                  boxShadow: "0 4px 12px 0 rgb(0 0 0 / 0.08)",
                  mb: 1,
                  p: 1,
                },
              },
            }}
          >
            <Box sx={{ px: 2, py: 1.5 }}>
              <Typography
                variant="caption"
                color="text.secondary"
                display="block"
              >
                Signed in as
              </Typography>
              <Typography variant="body2" fontWeight={600} noWrap>
                {user?.email}
              </Typography>
            </Box>

            <Divider sx={{ my: 1 }} />

            <MenuItem
              onClick={() => {
                handleCloseUserMenu();
                setSettingsOpen(true);
              }}
              sx={{
                borderRadius: 1.5,
                fontWeight: 600,
                fontSize: "0.875rem",
                py: 1,
              }}
            >
              Account Settings
            </MenuItem>

            <MenuItem
              onClick={handleSignOut}
              sx={{
                borderRadius: 1.5,
                color: "error.main",
                fontWeight: 600,
                fontSize: "0.875rem",
                py: 1,
              }}
            >
              Sign Out
            </MenuItem>
          </Menu>
        </Box>

        <UserSettingsModal
          open={settingsOpen}
          onClose={() => setSettingsOpen(false)}
        />
      </Drawer>

      {/* Main Content Container */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 4,
          minHeight: "100vh",
          backgroundColor: "background.default",
        }}
      >
        {children}
      </Box>
    </Box>
  );
}
