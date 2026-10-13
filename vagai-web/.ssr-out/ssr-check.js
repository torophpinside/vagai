import { toValue, ref, useModel, computed, onMounted, onUnmounted, mergeProps, mergeModels, useSSRContext, resolveComponent, unref, withCtx, createVNode, createTextVNode, openBlock, createBlock, createCommentVNode, toDisplayString, createSSRApp } from "vue";
import { ssrRenderAttrs, ssrRenderSlot, ssrRenderClass, ssrRenderComponent, ssrInterpolate, renderToString } from "vue/server-renderer";
import { useQuery, VueQueryPlugin } from "@tanstack/vue-query";
import axios from "axios";
import { useRouter } from "vue-router";
import { Target, LayoutDashboard, Package, Briefcase, CheckCircle2, Send, GraduationCap, FileText, BarChart3, FileEdit, Settings, CreditCard, LogOut, User, ChevronUp, ChevronDown } from "lucide-vue-next";
const API_URL = "/api";
const api = axios.create({ baseURL: API_URL });
let isRedirecting = false;
api.interceptors.response.use(
  (response) => response,
  (error) => {
    var _a;
    if (((_a = error.response) == null ? void 0 : _a.status) === 401 && !isRedirecting) {
      isRedirecting = true;
      localStorage.removeItem("vagai_token");
      localStorage.removeItem("vagai_user");
      delete api.defaults.headers.common["Authorization"];
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
const useStats = () => {
  return useQuery({ queryKey: ["stats"], queryFn: () => api.get("/stats").then((res) => res.data) });
};
const useAnalyticsMatching = (params = {}) => {
  return useQuery({
    queryKey: ["analytics", "matching", toValue(params)],
    queryFn: () => api.get("/v1/analytics/matching-distribution", { params: toValue(params) }).then((res) => res.data)
  });
};
const useAnalyticsResume = () => {
  return useQuery({
    queryKey: ["analytics", "resume"],
    queryFn: () => api.get("/v1/analytics/resume-insights").then((res) => res.data)
  });
};
const useAnalyticsJobTrends = (params = {}) => {
  return useQuery({
    queryKey: ["analytics", "job-trends", toValue(params)],
    queryFn: () => api.get("/v1/analytics/job-trends", { params: toValue(params) }).then((res) => res.data)
  });
};
const useAnalyticsUserMetrics = (userId) => {
  return useQuery({
    queryKey: ["analytics", "user-metrics", userId],
    queryFn: () => api.get(`/v1/analytics/user/${userId}`).then((res) => res.data),
    enabled: userId != null
  });
};
const useAnalyticsInterviewPrep = () => {
  return useQuery({
    queryKey: ["analytics", "interview-prep"],
    queryFn: () => api.get("/v1/analytics/interview-prep").then((res) => res.data)
  });
};
const useAnalyticsAIQuality = () => {
  return useQuery({
    queryKey: ["analytics", "ai-quality"],
    queryFn: () => api.get("/v1/analytics/ai-quality").then((res) => res.data)
  });
};
const useAnalyticsTimeToHire = () => {
  return useQuery({
    queryKey: ["analytics", "time-to-hire"],
    queryFn: () => api.get("/v1/analytics/time-to-hire").then((res) => res.data)
  });
};
const useAnalyticsSkillsGap = (params = {}) => {
  return useQuery({
    queryKey: ["analytics", "skills-gap", toValue(params)],
    queryFn: () => api.get("/v1/analytics/skills-gap", { params: toValue(params) }).then((res) => res.data)
  });
};
const useAnalyticsFunnel = () => {
  return useQuery({
    queryKey: ["analytics", "funnel"],
    queryFn: () => api.get("/v1/analytics/funnel").then((res) => res.data)
  });
};
const useAnalyticsGeographic = () => {
  return useQuery({
    queryKey: ["analytics", "geographic"],
    queryFn: () => api.get("/v1/analytics/geographic").then((res) => res.data)
  });
};
const useJobs = (filters = {}) => {
  return useQuery({
    queryKey: ["jobs", filters],
    queryFn: () => {
      const params = { ...toValue(filters) };
      if (Array.isArray(params.status)) {
        params.status = params.status.join(",");
      }
      if (Array.isArray(params.site)) {
        params.site = params.site.join(",");
      }
      return api.get("/jobs", { params }).then((res) => res.data);
    }
  });
};
const useMatches = (filters = {}) => {
  return useQuery({
    queryKey: ["matches", filters],
    queryFn: () => {
      const params = { ...toValue(filters) };
      if (Array.isArray(params.site) && params.site.length > 0) {
        params.site = params.site.join(",");
      }
      return api.get("/matches", { params }).then((res) => res.data);
    }
  });
};
const useSites = (options = {}) => {
  return useQuery({ queryKey: ["sites"], queryFn: () => api.get("/sites").then((res) => res.data), ...options });
};
const useResumeAnalyses = () => {
  return useQuery({ queryKey: ["resume-analyses"], queryFn: () => api.get("/resume-analyses").then((res) => res.data) });
};
const useInterviewPreps = () => {
  return useQuery({ queryKey: ["interview-preps"], queryFn: () => api.get("/interview-prep").then((res) => res.data) });
};
const useMe = () => {
  return useQuery({ queryKey: ["me"], queryFn: () => api.get("/me").then((res) => res.data) });
};
const usePlans = () => {
  return useQuery({ queryKey: ["plans"], queryFn: () => api.get("/plans").then((res) => res.data) });
};
const useAdminPlans = () => {
  return useQuery({ queryKey: ["plans", "admin"], queryFn: () => api.get("/admin/plans").then((res) => res.data) });
};
const useNegativeKeywords = (options = {}) => {
  return useQuery({ queryKey: ["negative-keywords"], queryFn: () => api.get("/negative-keywords").then((res) => res.data), ...options });
};
const listSites = () => api.get("/sites").then((res) => res.data);
const addSite = (site) => api.post("/sites", site);
const updateSite = (id, data) => api.patch(`/sites/${id}`, data);
const deleteSite = (id) => api.delete(`/sites/${id}`);
const updateProfile = (data) => api.patch("/me", data);
const updateNegativeKeywords = (keywords) => api.put("/negative-keywords", { keywords }).then((res) => res.data);
const createResume = (data) => api.post("/resumes", data);
const getResumeData = (id) => api.get(`/resumes/${id}/data`).then((res) => res.data);
const updateResumeData = (id, data) => api.put(`/resumes/${id}/data`, data);
const generateResumePDF = (id) => api.post(`/resumes/${id}/generate-pdf`, {}, { responseType: "blob" });
const listInterviewPreps = (filter = {}) => api.get("/interview-prep", { params: filter }).then((res) => res.data);
const getInterviewPrep = (id) => api.get(`/interview-prep/${id}`).then((res) => res.data);
const saveQuestionAnswer = (prepId, questionId, answer, selfAssessment, timeSpentSeconds = 0) => api.post(`/interview-prep/${prepId}/questions/${questionId}/answers`, { answer, self_assessment: selfAssessment, time_spent_seconds: timeSpentSeconds }).then((res) => res.data);
const user = ref(null);
const isAuthenticated = ref(false);
const TOKEN_KEY = "vagai_token";
const USER_KEY = "vagai_user";
const STALE_KEYS = ["vagai_resume_draft"];
function quotaError() {
  const err = new Error("Sessão autenticada, mas não foi possível salvá-la: armazenamento local cheio. Limpe os dados do site e tente novamente.");
  err.code = "QUOTA_EXCEEDED";
  return err;
}
function persistSession(token, userData) {
  const payload = JSON.stringify(userData);
  try {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, payload);
    return;
  } catch (err) {
    const cleaned = STALE_KEYS.filter((key) => {
      try {
        if (localStorage.getItem(key) !== null) {
          localStorage.removeItem(key);
          return true;
        }
      } catch {
      }
      return false;
    });
    if (cleaned.length > 0) {
      try {
        localStorage.setItem(TOKEN_KEY, token);
        localStorage.setItem(USER_KEY, payload);
        return;
      } catch {
      }
    }
    console.warn("Falha ao persistir sessão no localStorage:", err);
    throw quotaError();
  }
}
function clearStoredSession() {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  } catch {
  }
}
function initAuth$1() {
  let token = null;
  let rawUser = null;
  try {
    token = localStorage.getItem(TOKEN_KEY);
    rawUser = localStorage.getItem(USER_KEY);
  } catch {
    return;
  }
  if (!token || !rawUser) return;
  try {
    user.value = JSON.parse(rawUser);
    isAuthenticated.value = true;
    api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  } catch {
    clearStoredSession();
  }
}
async function login(email, password) {
  const response = await api.post("/auth/login", { email, password });
  const { token, user: userData } = response.data;
  persistSession(token, userData);
  api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  user.value = userData;
  isAuthenticated.value = true;
  return response.data;
}
async function register(name, email, password) {
  const response = await api.post("/auth/register", { name, email, password });
  const { token, user: userData } = response.data;
  persistSession(token, userData);
  api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  user.value = userData;
  isAuthenticated.value = true;
  return response.data;
}
function logout() {
  clearStoredSession();
  delete api.defaults.headers.common["Authorization"];
  user.value = null;
  isAuthenticated.value = false;
}
function updateUser(userData) {
  const current = user.value || {};
  const updated = { ...current, ...userData };
  try {
    localStorage.setItem(USER_KEY, JSON.stringify(updated));
  } catch (err) {
    console.warn("Não foi possível persistir o perfil atualizado:", err);
  }
  user.value = updated;
}
function useAuth() {
  return {
    user,
    isAuthenticated,
    initAuth: initAuth$1,
    login,
    register,
    logout,
    updateUser
  };
}
const _export_sfc = (sfc, props) => {
  const target = sfc.__vccOpts || sfc;
  for (const [key, val] of props) {
    target[key] = val;
  }
  return target;
};
const _sfc_main$1 = {
  __name: "DropdownMenu",
  __ssrInlineRender: true,
  props: /* @__PURE__ */ mergeModels({
    position: {
      type: String,
      default: "bottom-left",
      validator: (v) => ["bottom-left", "bottom-right", "top-left", "top-right"].includes(v)
    },
    width: {
      type: String,
      default: "w-56"
    },
    panelClass: {
      type: String,
      default: ""
    }
  }, {
    "open": { default: false },
    "openModifiers": {}
  }),
  emits: ["update:open"],
  setup(__props) {
    const props = __props;
    const isOpen = useModel(__props, "open");
    const containerRef = ref(null);
    const positionClasses = {
      "bottom-left": "left-0 mt-2",
      "bottom-right": "right-0 mt-2",
      "top-left": "left-0 mb-2 bottom-full",
      "top-right": "right-0 mb-2 bottom-full"
    };
    const panelClasses = computed(() => {
      const pos = positionClasses[props.position];
      const w = props.width || "";
      return `absolute ${pos} ${w} glass-card p-2 z-50 overflow-hidden ${props.panelClass}`.trim();
    });
    function toggle() {
      isOpen.value = !isOpen.value;
    }
    function close() {
      isOpen.value = false;
    }
    function onClickOutside(e) {
      if (containerRef.value && !containerRef.value.contains(e.target)) {
        isOpen.value = false;
      }
    }
    onMounted(() => document.addEventListener("click", onClickOutside));
    onUnmounted(() => document.removeEventListener("click", onClickOutside));
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({
        class: "relative",
        ref_key: "containerRef",
        ref: containerRef
      }, _attrs))} data-v-c1d23830>`);
      ssrRenderSlot(_ctx.$slots, "trigger", {
        open: isOpen.value,
        toggle
      }, null, _push, _parent);
      if (isOpen.value) {
        _push(`<div class="${ssrRenderClass(panelClasses.value)}" data-v-c1d23830>`);
        ssrRenderSlot(_ctx.$slots, "default", { close }, null, _push, _parent);
        _push(`</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$1 = _sfc_main$1.setup;
_sfc_main$1.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/DropdownMenu.vue");
  return _sfc_setup$1 ? _sfc_setup$1(props, ctx) : void 0;
};
const DropdownMenu = /* @__PURE__ */ _export_sfc(_sfc_main$1, [["__scopeId", "data-v-c1d23830"]]);
const _sfc_main = {
  __name: "App",
  __ssrInlineRender: true,
  setup(__props) {
    const router = useRouter();
    const { user: user2, isAuthenticated: isAuthenticated2, initAuth: initAuth2, logout: logout2 } = useAuth();
    const showUserMenu = ref(false);
    onMounted(() => {
      initAuth2();
    });
    const userName = computed(() => {
      var _a;
      return ((_a = user2.value) == null ? void 0 : _a.name) || "Usuário";
    });
    const firstName = computed(() => {
      var _a, _b;
      return ((_b = (_a = user2.value) == null ? void 0 : _a.name) == null ? void 0 : _b.split(" ")[0]) || "Usuário";
    });
    const userPlan = computed(() => {
      var _a, _b;
      const plan = ((_a = user2.value) == null ? void 0 : _a.plan) || ((_b = user2.value) == null ? void 0 : _b.plan_slug) || "free";
      return plan === "free" ? "Plano Free" : `Plano ${plan.charAt(0).toUpperCase()}${plan.slice(1)}`;
    });
    const isAdmin = computed(() => {
      var _a;
      return ((_a = user2.value) == null ? void 0 : _a.role) === "admin";
    });
    function handleLogout() {
      logout2();
      router.push("/login");
    }
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_view = resolveComponent("router-view");
      const _component_router_link = resolveComponent("router-link");
      if (!unref(isAuthenticated2)) {
        _push(`<div${ssrRenderAttrs(mergeProps({ class: "min-h-screen bg-[#0f172a]" }, _attrs))}>`);
        _push(ssrRenderComponent(_component_router_view, null, null, _parent));
        _push(`</div>`);
      } else {
        _push(`<div${ssrRenderAttrs(mergeProps({ class: "flex min-h-screen bg-[#0f172a] text-slate-200 overflow-hidden" }, _attrs))}><aside class="w-72 bg-slate-900/50 border-r border-white/5 backdrop-blur-xl flex flex-col"><div class="p-8"><div class="flex items-center gap-3 mb-12"><div class="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">`);
        _push(ssrRenderComponent(unref(Target), { class: "text-white w-6 h-6" }, null, _parent));
        _push(`</div><span class="text-2xl font-bold font-outfit text-white tracking-tight">VagAI</span></div><nav class="space-y-2">`);
        _push(ssrRenderComponent(_component_router_link, {
          to: "/",
          class: "nav-link",
          "active-class": "active"
        }, {
          default: withCtx((_, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(ssrRenderComponent(unref(LayoutDashboard), { class: "w-5 h-5" }, null, _parent2, _scopeId));
              _push2(`<span${_scopeId}>Dashboard</span>`);
            } else {
              return [
                createVNode(unref(LayoutDashboard), { class: "w-5 h-5" }),
                createVNode("span", null, "Dashboard")
              ];
            }
          }),
          _: 1
        }, _parent));
        if (isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/admin/plans",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(Package), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Planos</span>`);
              } else {
                return [
                  createVNode(unref(Package), { class: "w-5 h-5" }),
                  createVNode("span", null, "Planos")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/jobs",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(Briefcase), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Vagas</span>`);
              } else {
                return [
                  createVNode(unref(Briefcase), { class: "w-5 h-5" }),
                  createVNode("span", null, "Vagas")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/matches",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(CheckCircle2), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Matches</span>`);
              } else {
                return [
                  createVNode(unref(CheckCircle2), { class: "w-5 h-5" }),
                  createVNode("span", null, "Matches")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/applied",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(Send), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Candidatadas</span>`);
              } else {
                return [
                  createVNode(unref(Send), { class: "w-5 h-5" }),
                  createVNode("span", null, "Candidatadas")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/interview-prep",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(GraduationCap), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Preparação</span>`);
              } else {
                return [
                  createVNode(unref(GraduationCap), { class: "w-5 h-5" }),
                  createVNode("span", null, "Preparação")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/analysis",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(FileText), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Análise</span>`);
              } else {
                return [
                  createVNode(unref(FileText), { class: "w-5 h-5" }),
                  createVNode("span", null, "Análise")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/analytics",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(BarChart3), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Analytics</span>`);
              } else {
                return [
                  createVNode(unref(BarChart3), { class: "w-5 h-5" }),
                  createVNode("span", null, "Analytics")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        if (!isAdmin.value) {
          _push(ssrRenderComponent(_component_router_link, {
            to: "/resume-editor/new",
            class: "nav-link",
            "active-class": "active"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(FileEdit), { class: "w-5 h-5" }, null, _parent2, _scopeId));
                _push2(`<span${_scopeId}>Currículo</span>`);
              } else {
                return [
                  createVNode(unref(FileEdit), { class: "w-5 h-5" }),
                  createVNode("span", null, "Currículo")
                ];
              }
            }),
            _: 1
          }, _parent));
        } else {
          _push(`<!---->`);
        }
        _push(ssrRenderComponent(_component_router_link, {
          to: "/settings",
          class: "nav-link",
          "active-class": "active"
        }, {
          default: withCtx((_, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(ssrRenderComponent(unref(Settings), { class: "w-5 h-5" }, null, _parent2, _scopeId));
              _push2(`<span${_scopeId}>Configurações</span>`);
            } else {
              return [
                createVNode(unref(Settings), { class: "w-5 h-5" }),
                createVNode("span", null, "Configurações")
              ];
            }
          }),
          _: 1
        }, _parent));
        _push(`</nav></div><div class="mt-auto p-8">`);
        _push(ssrRenderComponent(DropdownMenu, {
          open: showUserMenu.value,
          "onUpdate:open": ($event) => showUserMenu.value = $event,
          position: "top-left",
          width: "",
          "panel-class": "right-0"
        }, {
          trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(`<div class="glass-card p-4 flex items-center gap-4"${_scopeId}><div class="w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(User), { class: "w-6 h-6 text-indigo-400" }, null, _parent2, _scopeId));
              _push2(`</div><div class="flex-1 min-w-0"${_scopeId}><div class="text-sm font-bold text-white truncate"${_scopeId}>${ssrInterpolate(userName.value)}</div><div class="text-xs text-slate-500 truncate"${_scopeId}>${ssrInterpolate(userPlan.value)}</div></div><button class="text-slate-400 hover:text-white"${_scopeId}>`);
              if (open) {
                _push2(ssrRenderComponent(unref(ChevronUp), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              } else {
                _push2(ssrRenderComponent(unref(ChevronDown), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              }
              _push2(`</button></div>`);
            } else {
              return [
                createVNode("div", {
                  class: "glass-card p-4 flex items-center gap-4",
                  onClick: toggle
                }, [
                  createVNode("div", { class: "w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center" }, [
                    createVNode(unref(User), { class: "w-6 h-6 text-indigo-400" })
                  ]),
                  createVNode("div", { class: "flex-1 min-w-0" }, [
                    createVNode("div", { class: "text-sm font-bold text-white truncate" }, toDisplayString(userName.value), 1),
                    createVNode("div", { class: "text-xs text-slate-500 truncate" }, toDisplayString(userPlan.value), 1)
                  ]),
                  createVNode("button", { class: "text-slate-400 hover:text-white" }, [
                    open ? (openBlock(), createBlock(unref(ChevronUp), {
                      key: 0,
                      class: "w-4 h-4"
                    })) : (openBlock(), createBlock(unref(ChevronDown), {
                      key: 1,
                      class: "w-4 h-4"
                    }))
                  ])
                ], 8, ["onClick"])
              ];
            }
          }),
          default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(ssrRenderComponent(_component_router_link, {
                onClick: close,
                to: "/settings",
                class: "block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg"
              }, {
                default: withCtx((_, _push3, _parent3, _scopeId2) => {
                  if (_push3) {
                    _push3(ssrRenderComponent(unref(Settings), { class: "w-4 h-4 inline mr-2" }, null, _parent3, _scopeId2));
                    _push3(` Configurações `);
                  } else {
                    return [
                      createVNode(unref(Settings), { class: "w-4 h-4 inline mr-2" }),
                      createTextVNode(" Configurações ")
                    ];
                  }
                }),
                _: 2
              }, _parent2, _scopeId));
              if (!isAdmin.value) {
                _push2(ssrRenderComponent(_component_router_link, {
                  onClick: close,
                  to: "/settings/billing",
                  class: "block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg"
                }, {
                  default: withCtx((_, _push3, _parent3, _scopeId2) => {
                    if (_push3) {
                      _push3(ssrRenderComponent(unref(CreditCard), { class: "w-4 h-4 inline mr-2" }, null, _parent3, _scopeId2));
                      _push3(` Billing `);
                    } else {
                      return [
                        createVNode(unref(CreditCard), { class: "w-4 h-4 inline mr-2" }),
                        createTextVNode(" Billing ")
                      ];
                    }
                  }),
                  _: 2
                }, _parent2, _scopeId));
              } else {
                _push2(`<!---->`);
              }
              _push2(`<button class="w-full text-left px-3 py-2 text-sm text-red-400 hover:bg-white/5 rounded-lg"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(LogOut), { class: "w-4 h-4 inline mr-2" }, null, _parent2, _scopeId));
              _push2(` Sair </button>`);
            } else {
              return [
                createVNode(_component_router_link, {
                  onClick: close,
                  to: "/settings",
                  class: "block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg"
                }, {
                  default: withCtx(() => [
                    createVNode(unref(Settings), { class: "w-4 h-4 inline mr-2" }),
                    createTextVNode(" Configurações ")
                  ]),
                  _: 1
                }, 8, ["onClick"]),
                !isAdmin.value ? (openBlock(), createBlock(_component_router_link, {
                  key: 0,
                  onClick: close,
                  to: "/settings/billing",
                  class: "block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg"
                }, {
                  default: withCtx(() => [
                    createVNode(unref(CreditCard), { class: "w-4 h-4 inline mr-2" }),
                    createTextVNode(" Billing ")
                  ]),
                  _: 1
                }, 8, ["onClick"])) : createCommentVNode("", true),
                createVNode("button", {
                  onClick: handleLogout,
                  class: "w-full text-left px-3 py-2 text-sm text-red-400 hover:bg-white/5 rounded-lg"
                }, [
                  createVNode(unref(LogOut), { class: "w-4 h-4 inline mr-2" }),
                  createTextVNode(" Sair ")
                ])
              ];
            }
          }),
          _: 1
        }, _parent));
        _push(`</div></aside><main class="flex-1 overflow-y-auto bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950/20"><header class="h-20 border-b border-white/5 flex items-center justify-between px-12 backdrop-blur-md sticky top-0 z-10"><div class="text-slate-400 text-sm"> Bem-vindo de volta, <span class="text-white font-bold">${ssrInterpolate(firstName.value)}</span></div><div class="flex items-center gap-6"><div class="w-[1px] h-6 bg-white/10"></div><div class="text-right"><div class="text-xs text-slate-500">Status do Scanner</div><div class="text-xs text-emerald-400 font-bold flex items-center gap-1"><span class="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse"></span> Ativo </div></div></div></header><div class="p-12 animate-in">`);
        _push(ssrRenderComponent(_component_router_view, null, null, _parent));
        _push(`</div></main></div>`);
      }
    };
  }
};
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/App.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};
const locationStub = {
  pathname: "/",
  search: "",
  hash: "",
  host: "localhost",
  hostname: "localhost",
  protocol: "http:",
  port: "",
  origin: "http://localhost",
  href: "http://localhost/",
  replace() {
  },
  assign() {
  }
};
globalThis.localStorage = {
  getItem(key) {
    if (key === "vagai_token") return "ssr-token";
    if (key === "vagai_user")
      return JSON.stringify({ id: 1, name: "SSR Tester", role: process.env.ROLE || "user", plan: "free" });
    return null;
  },
  setItem() {
  },
  removeItem() {
  }
};
globalThis.window = {
  location: locationStub,
  history: { state: { forward: [], back: [] }, pushState() {
  }, replaceState() {
  }, go() {
  }, back() {
  } },
  addEventListener() {
  },
  removeEventListener() {
  },
  matchMedia: () => ({ matches: false, addEventListener() {
  }, removeEventListener() {
  } }),
  navigator: { userAgent: "ssr" },
  document: void 0
};
globalThis.document = {
  title: "",
  addEventListener() {
  },
  removeEventListener() {
  },
  querySelector: () => null,
  querySelectorAll: () => [],
  visibilityState: "visible",
  hidden: false,
  baseURI: "http://localhost/"
};
Object.defineProperty(globalThis, "navigator", {
  value: { userAgent: "ssr" },
  configurable: true,
  writable: true
});
Object.defineProperty(globalThis, "location", {
  value: locationStub,
  configurable: true,
  writable: true
});
Object.defineProperty(globalThis, "history", {
  value: globalThis.window.history,
  configurable: true,
  writable: true
});
window.document = globalThis.document;
const { initAuth } = useAuth();
initAuth();
console.error("[ssr-check] iniciando, ROLE =", process.env.ROLE, "NAV_TO =", process.env.NAV_TO);
setTimeout(() => {
  console.error("SSR TIMEOUT: navegação nunca completou");
  process.exit(2);
}, 8e3);
(async () => {
  try {
    const { default: router } = await import("./assets/router-CBtgVjnp.js");
    const app = createSSRApp(_sfc_main);
    app.use(VueQueryPlugin);
    app.use(router);
    const target = process.env.NAV_TO || "/";
    await router.push(target);
    const finalPath = router.currentRoute.value.path;
    let renderOk = true;
    let navItems = null;
    let hrefs = [];
    let pageMarkers = null;
    try {
      const html = await renderToString(app);
      const navMatch = html.match(/<nav[\s\S]*?<\/nav>/);
      hrefs = navMatch ? [...navMatch[0].matchAll(/href="([^"]*)"/g)].map((m) => m[1]) : [];
      navItems = hrefs.length;
      const markers = ["Billing", "Adicionar Fonte", "Fontes Cadastradas", "Palavras-chave de bloqueio", "Seu Perfil"];
      pageMarkers = Object.fromEntries(markers.map((m) => [m, html.includes(m)]));
    } catch (e) {
      renderOk = false;
      hrefs = [`<render-error: ${e && e.message ? e.message : e}>`];
    }
    console.log(
      JSON.stringify({
        role: process.env.ROLE || "user",
        nav_to: target,
        final_path: finalPath,
        redirected: finalPath !== target,
        nav_items: navItems,
        hrefs,
        page_markers: pageMarkers,
        render_ok: renderOk
      })
    );
    process.exit(0);
  } catch (e) {
    console.error("SSR ERROR:", e && e.stack ? e.stack : e);
    process.exit(1);
  }
})();
export {
  useAnalyticsUserMetrics as A,
  useAnalyticsInterviewPrep as B,
  useAnalyticsAIQuality as C,
  DropdownMenu as D,
  useAnalyticsTimeToHire as E,
  useAnalyticsSkillsGap as F,
  useAnalyticsFunnel as G,
  useAnalyticsGeographic as H,
  useAdminPlans as I,
  useSites as a,
  useJobs as b,
  useMatches as c,
  listInterviewPreps as d,
  useResumeAnalyses as e,
  useMe as f,
  useNegativeKeywords as g,
  updateProfile as h,
  addSite as i,
  deleteSite as j,
  updateSite as k,
  listSites as l,
  updateNegativeKeywords as m,
  useAuth as n,
  usePlans as o,
  getResumeData as p,
  generateResumePDF as q,
  updateResumeData as r,
  createResume as s,
  getInterviewPrep as t,
  useStats as u,
  saveQuestionAnswer as v,
  useInterviewPreps as w,
  useAnalyticsMatching as x,
  useAnalyticsResume as y,
  useAnalyticsJobTrends as z
};
