import { useRouter, useRoute, createRouter, createWebHistory } from "vue-router";
import { computed, mergeProps, unref, useSSRContext, reactive, ref, watch, resolveComponent, withCtx, openBlock, createBlock, Fragment, renderList, createVNode, createCommentVNode, createTextVNode, toDisplayString, onMounted, onUnmounted } from "vue";
import { ssrRenderAttrs, ssrInterpolate, ssrRenderComponent, ssrRenderAttr, ssrRenderList, ssrRenderClass, ssrIncludeBooleanAttr, ssrRenderStyle, ssrLooseContain, ssrLooseEqual, ssrRenderSlot } from "vue/server-renderer";
import { u as useStats, a as useSites, b as useJobs, D as DropdownMenu, l as listSites, c as useMatches, d as listInterviewPreps, e as useResumeAnalyses, f as useMe, g as useNegativeKeywords, h as updateProfile, i as addSite, j as deleteSite, k as updateSite, m as updateNegativeKeywords, n as useAuth, o as usePlans, p as getResumeData, q as generateResumePDF, r as updateResumeData, s as createResume, t as getInterviewPrep, v as saveQuestionAnswer, w as useInterviewPreps, x as useAnalyticsMatching, y as useAnalyticsResume, z as useAnalyticsJobTrends, A as useAnalyticsUserMetrics, B as useAnalyticsInterviewPrep, C as useAnalyticsAIQuality, E as useAnalyticsTimeToHire, F as useAnalyticsSkillsGap, G as useAnalyticsFunnel, H as useAnalyticsGeographic } from "../ssr-check.js";
import { Bar, Doughnut } from "vue-chartjs";
import { Briefcase, TrendingUp, CheckCircle2, Target, Send, Rocket, Globe, Activity, Code2, Cpu, PieChart, Layers, Search, X, Filter, ChevronDown, Plus, RefreshCcw, ExternalLink, Trash2, ChevronLeft, ChevronRight, ArrowLeft, Loader2, RefreshCw, ArrowUpDown, Building2, Calendar, EyeOff, CheckCircle, Paperclip, Sparkles, BrainCircuit, GraduationCap, Upload, FileText, Loader, ThumbsUp, AlertTriangle, AlertCircle, Lightbulb, History, FileX, MapPin, CreditCard, Power, Ban, Check, User, Wrench, Award, Save, Download, CircleCheck, CircleAlert, FileSearch, Eye, Wand2, Play, Inbox, Timer, Puzzle } from "lucide-vue-next";
import { Chart, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement, LineElement, PointElement } from "chart.js";
import { useQueryClient, useMutation } from "@tanstack/vue-query";
import "axios";
const _sfc_main$x = {
  __name: "Dashboard",
  __ssrInlineRender: true,
  setup(__props) {
    Chart.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement);
    const { data: stats, isLoading } = useStats();
    const chartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false }
      },
      scales: {
        y: { beginAtZero: true }
      }
    };
    const pieOptions2 = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: "right" }
      }
    };
    const langColors = ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16", "#f97316", "#6366f1"];
    const langChartData = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = stats.value) == null ? void 0 : _a.languages) || []).map((l) => l.name),
        datasets: [{
          label: "Vagas",
          data: (((_b = stats.value) == null ? void 0 : _b.languages) || []).map((l) => l.count),
          backgroundColor: langColors,
          borderRadius: 4
        }]
      };
    });
    const kwChartData = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = stats.value) == null ? void 0 : _a.keywords) || []).map((k) => k.name),
        datasets: [{
          label: "Vagas",
          data: (((_b = stats.value) == null ? void 0 : _b.keywords) || []).map((k) => k.count),
          backgroundColor: langColors,
          borderRadius: 4
        }]
      };
    });
    const langPieData = computed(() => {
      var _a;
      const langs = (((_a = stats.value) == null ? void 0 : _a.languages) || []).slice(0, 5);
      return {
        labels: langs.map((l) => l.name),
        datasets: [{
          data: langs.map((l) => l.count),
          backgroundColor: langColors.slice(0, 5)
        }]
      };
    });
    const kwPieData = computed(() => {
      var _a;
      const kws = (((_a = stats.value) == null ? void 0 : _a.keywords) || []).slice(0, 5);
      return {
        labels: kws.map((k) => k.name),
        datasets: [{
          data: kws.map((k) => k.count),
          backgroundColor: langColors.slice(0, 5)
        }]
      };
    });
    return (_ctx, _push, _parent, _attrs) => {
      var _a, _b, _c, _d;
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-12" }, _attrs))}><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Dashboard</h1><p class="text-slate-400">Visão geral de suas vagas.</p></div>`);
      if (unref(isLoading)) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else {
        _push(`<div class="space-y-12"><div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8"><div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default"><div class="flex items-start justify-between"><div><div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Total de Vagas</div><div class="text-5xl font-bold text-white font-outfit">${ssrInterpolate(((_a = unref(stats)) == null ? void 0 : _a.total_jobs) || 0)}</div></div><div class="w-12 h-12 bg-blue-500/10 rounded-2xl flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform duration-500">`);
        _push(ssrRenderComponent(unref(Briefcase), { class: "w-6 h-6" }, null, _parent));
        _push(`</div></div><div class="mt-6 flex items-center gap-2 text-xs text-emerald-400 font-bold uppercase">`);
        _push(ssrRenderComponent(unref(TrendingUp), { class: "w-4 h-4" }, null, _parent));
        _push(`<span>+12% este mês</span></div></div><div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default"><div class="flex items-start justify-between"><div><div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Matches Ideais</div><div class="text-5xl font-bold text-white font-outfit">${ssrInterpolate(((_b = unref(stats)) == null ? void 0 : _b.total_matches) || 0)}</div></div><div class="w-12 h-12 bg-indigo-500/10 rounded-2xl flex items-center justify-center text-indigo-400 group-hover:scale-110 transition-transform duration-500">`);
        _push(ssrRenderComponent(unref(CheckCircle2), { class: "w-6 h-6" }, null, _parent));
        _push(`</div></div><div class="mt-6 flex items-center gap-2 text-xs text-amber-400 font-bold uppercase">`);
        _push(ssrRenderComponent(unref(Target), { class: "w-4 h-4" }, null, _parent));
        _push(`<span>Alta Similaridade</span></div></div><div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default"><div class="flex items-start justify-between"><div><div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Candidatadas</div><div class="text-5xl font-bold text-white font-outfit">${ssrInterpolate(((_c = unref(stats)) == null ? void 0 : _c.total_applied) || 0)}</div></div><div class="w-12 h-12 bg-emerald-500/10 rounded-2xl flex items-center justify-center text-emerald-400 group-hover:scale-110 transition-transform duration-500">`);
        _push(ssrRenderComponent(unref(Send), { class: "w-6 h-6" }, null, _parent));
        _push(`</div></div><div class="mt-6 flex items-center gap-2 text-xs text-emerald-400 font-bold uppercase">`);
        _push(ssrRenderComponent(unref(Rocket), { class: "w-4 h-4" }, null, _parent));
        _push(`<span>BOA sorte!</span></div></div><div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default"><div class="flex items-start justify-between"><div><div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Fontes Ativas</div><div class="text-5xl font-bold text-white font-outfit">${ssrInterpolate(((_d = unref(stats)) == null ? void 0 : _d.active_sites) || 0)}</div></div><div class="w-12 h-12 bg-purple-500/10 rounded-2xl flex items-center justify-center text-purple-400 group-hover:scale-110 transition-transform duration-500">`);
        _push(ssrRenderComponent(unref(Globe), { class: "w-6 h-6" }, null, _parent));
        _push(`</div></div><div class="mt-6 flex items-center gap-2 text-xs text-slate-500 font-bold uppercase">`);
        _push(ssrRenderComponent(unref(Activity), { class: "w-4 h-4" }, null, _parent));
        _push(`<span>Scanner Saudável</span></div></div></div><div class="grid grid-cols-1 lg:grid-cols-2 gap-12"><div class="glass-card p-10"><div class="flex items-center justify-between mb-10"><h2 class="text-2xl font-bold text-white font-outfit">Linguagens</h2>`);
        _push(ssrRenderComponent(unref(Code2), { class: "text-slate-500 w-6 h-6" }, null, _parent));
        _push(`</div><div class="h-[400px]">`);
        _push(ssrRenderComponent(unref(Bar), {
          data: langChartData.value,
          options: chartOptions
        }, null, _parent));
        _push(`</div></div><div class="glass-card p-10"><div class="flex items-center justify-between mb-10"><h2 class="text-2xl font-bold text-white font-outfit">Tecnologias</h2>`);
        _push(ssrRenderComponent(unref(Cpu), { class: "text-slate-500 w-6 h-6" }, null, _parent));
        _push(`</div><div class="h-[400px]">`);
        _push(ssrRenderComponent(unref(Bar), {
          data: kwChartData.value,
          options: chartOptions
        }, null, _parent));
        _push(`</div></div></div><div class="grid grid-cols-1 lg:grid-cols-2 gap-12"><div class="glass-card p-10"><div class="flex items-center justify-between mb-10"><h2 class="text-2xl font-bold text-white font-outfit">Market Share (Linguagens)</h2>`);
        _push(ssrRenderComponent(unref(PieChart), { class: "text-slate-500 w-6 h-6" }, null, _parent));
        _push(`</div><div class="h-[300px]">`);
        _push(ssrRenderComponent(unref(Doughnut), {
          data: langPieData.value,
          options: pieOptions2
        }, null, _parent));
        _push(`</div></div><div class="glass-card p-10"><div class="flex items-center justify-between mb-10"><h2 class="text-2xl font-bold text-white font-outfit">Market Share (Tech)</h2>`);
        _push(ssrRenderComponent(unref(Layers), { class: "text-slate-500 w-6 h-6" }, null, _parent));
        _push(`</div><div class="h-[300px]">`);
        _push(ssrRenderComponent(unref(Doughnut), {
          data: kwPieData.value,
          options: pieOptions2
        }, null, _parent));
        _push(`</div></div></div></div>`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$x = _sfc_main$x.setup;
_sfc_main$x.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Dashboard.vue");
  return _sfc_setup$x ? _sfc_setup$x(props, ctx) : void 0;
};
const _sfc_main$w = {
  __name: "Jobs",
  __ssrInlineRender: true,
  setup(__props) {
    useQueryClient();
    const { data: sites } = useSites();
    const filters = reactive({
      site: [],
      status: [],
      search: "",
      page: 1,
      limit: 20
    });
    const searchQuery = ref("");
    let searchTimer;
    watch(searchQuery, (val) => {
      clearTimeout(searchTimer);
      searchTimer = setTimeout(() => {
        filters.search = val;
      }, 300);
    });
    watch([() => filters.site, () => filters.status, () => filters.search], () => {
      filters.page = 1;
    }, { deep: true });
    const { data: jobsResponse, isLoading } = useJobs(filters);
    const jobs = computed(() => {
      var _a;
      return ((_a = jobsResponse.value) == null ? void 0 : _a.data) || [];
    });
    const total = computed(() => {
      var _a;
      return ((_a = jobsResponse.value) == null ? void 0 : _a.total) || 0;
    });
    const totalPages = computed(() => {
      var _a;
      return ((_a = jobsResponse.value) == null ? void 0 : _a.totalPages) || 0;
    });
    const showingInfo = computed(() => {
      if (total.value === 0) return "Nenhuma vaga encontrada";
      const from = (filters.page - 1) * filters.limit + 1;
      const to = Math.min(filters.page * filters.limit, total.value);
      return `Mostrando ${from} a ${to} de ${total.value} vagas`;
    });
    const statusOptions = [
      { value: "new", label: "Novas" },
      { value: "matched", label: "Match" },
      { value: "analyzed", label: "Analisadas" },
      { value: "unmatched", label: "Sem Match" },
      { value: "ignored", label: "Ignoradas" }
    ];
    const statusLabelMap = {
      new: "Nova",
      matched: "Match",
      analyzed: "Analisada",
      unmatched: "Sem Match",
      ignored: "Ignorada"
    };
    const siteOpen = ref(false);
    const toggleSite = (id) => {
      const idx = filters.site.indexOf(id);
      if (idx >= 0) {
        filters.site.splice(idx, 1);
      } else {
        filters.site.push(id);
      }
    };
    const statusOpen = ref(false);
    const toggleStatus = (value) => {
      const idx = filters.status.indexOf(value);
      if (idx >= 0) {
        filters.status.splice(idx, 1);
      } else {
        filters.status.push(value);
      }
    };
    const statusClass = (status) => {
      const classes = {
        new: "bg-indigo-500/10 text-indigo-400 border-indigo-500/20",
        matched: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
        analyzed: "bg-amber-500/10 text-amber-400 border-amber-500/20",
        unmatched: "bg-orange-500/10 text-orange-400 border-orange-500/20",
        ignored: "bg-slate-700/50 text-slate-400 border-white/5"
      };
      return classes[status] || classes.new;
    };
    const formatDate = (date) => {
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit"
      }).format(new Date(date));
    };
    const scoreClass = (score) => {
      if (score >= 0.8) return "bg-emerald-500/10 text-emerald-400";
      if (score >= 0.6) return "bg-indigo-500/10 text-indigo-400";
      if (score >= 0.4) return "bg-amber-500/10 text-amber-400";
      return "bg-slate-500/10 text-slate-400";
    };
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="flex items-center justify-between"><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Vagas</h1><p class="text-slate-400">Gerencie as oportunidades encontradas pelo scanner.</p></div><div class="flex gap-4 items-center"><div class="relative flex-1 max-w-xs">`);
      _push(ssrRenderComponent(unref(Search), { class: "absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" }, null, _parent));
      _push(`<input${ssrRenderAttr("value", searchQuery.value)} type="text" placeholder="Buscar por título..." class="input-field pl-10 h-12 w-full"></div>`);
      _push(ssrRenderComponent(DropdownMenu, {
        open: siteOpen.value,
        "onUpdate:open": ($event) => siteOpen.value = $event
      }, {
        trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"${_scopeId}>`);
            _push2(ssrRenderComponent(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }, null, _parent2, _scopeId));
            _push2(`<span class="text-sm text-slate-300"${_scopeId}>`);
            if (filters.site.length === 0) {
              _push2(`<!--[-->Todos os Sites<!--]-->`);
            } else {
              _push2(`<!--[-->${ssrInterpolate(filters.site.length)} site${ssrInterpolate(filters.site.length > 1 ? "s" : "")}<!--]-->`);
            }
            _push2(`</span>`);
            _push2(ssrRenderComponent(unref(ChevronDown), {
              class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
            }, null, _parent2, _scopeId));
            _push2(`</button>`);
          } else {
            return [
              createVNode("button", {
                onClick: toggle,
                class: "flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"
              }, [
                createVNode(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }),
                createVNode("span", { class: "text-sm text-slate-300" }, [
                  filters.site.length === 0 ? (openBlock(), createBlock(Fragment, { key: 0 }, [
                    createTextVNode("Todos os Sites")
                  ], 64)) : (openBlock(), createBlock(Fragment, { key: 1 }, [
                    createTextVNode(toDisplayString(filters.site.length) + " site" + toDisplayString(filters.site.length > 1 ? "s" : ""), 1)
                  ], 64))
                ]),
                createVNode(unref(ChevronDown), {
                  class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
                }, null, 8, ["class"])
              ], 8, ["onClick"])
            ];
          }
        }),
        default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<!--[-->`);
            ssrRenderList(unref(sites), (site) => {
              _push2(`<button class="${ssrRenderClass([filters.site.includes(site.id) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}><div class="${ssrRenderClass([filters.site.includes(site.id) ? "bg-indigo-500 border-indigo-500" : "border-slate-600", "w-4 h-4 rounded border flex items-center justify-center transition-all"])}"${_scopeId}>`);
              if (filters.site.includes(site.id)) {
                _push2(`<svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"${_scopeId}><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"${_scopeId}></path></svg>`);
              } else {
                _push2(`<!---->`);
              }
              _push2(`</div> ${ssrInterpolate(site.name)}</button>`);
            });
            _push2(`<!--]-->`);
            if (filters.site.length > 0) {
              _push2(`<div class="border-t border-white/5 p-2"${_scopeId}><button class="flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(X), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              _push2(` Limpar filtros </button></div>`);
            } else {
              _push2(`<!---->`);
            }
          } else {
            return [
              (openBlock(true), createBlock(Fragment, null, renderList(unref(sites), (site) => {
                return openBlock(), createBlock("button", {
                  key: site.id,
                  onClick: ($event) => toggleSite(site.id),
                  class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", filters.site.includes(site.id) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
                }, [
                  createVNode("div", {
                    class: ["w-4 h-4 rounded border flex items-center justify-center transition-all", filters.site.includes(site.id) ? "bg-indigo-500 border-indigo-500" : "border-slate-600"]
                  }, [
                    filters.site.includes(site.id) ? (openBlock(), createBlock("svg", {
                      key: 0,
                      class: "w-3 h-3 text-white",
                      fill: "none",
                      viewBox: "0 0 24 24",
                      stroke: "currentColor"
                    }, [
                      createVNode("path", {
                        "stroke-linecap": "round",
                        "stroke-linejoin": "round",
                        "stroke-width": "3",
                        d: "M5 13l4 4L19 7"
                      })
                    ])) : createCommentVNode("", true)
                  ], 2),
                  createTextVNode(" " + toDisplayString(site.name), 1)
                ], 10, ["onClick"]);
              }), 128)),
              filters.site.length > 0 ? (openBlock(), createBlock("div", {
                key: 0,
                class: "border-t border-white/5 p-2"
              }, [
                createVNode("button", {
                  onClick: ($event) => {
                    filters.site = [];
                    close();
                  },
                  class: "flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"
                }, [
                  createVNode(unref(X), { class: "w-4 h-4" }),
                  createTextVNode(" Limpar filtros ")
                ], 8, ["onClick"])
              ])) : createCommentVNode("", true)
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(DropdownMenu, {
        open: statusOpen.value,
        "onUpdate:open": ($event) => statusOpen.value = $event,
        position: "bottom-right"
      }, {
        trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"${_scopeId}>`);
            _push2(ssrRenderComponent(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }, null, _parent2, _scopeId));
            _push2(`<span class="text-sm text-slate-300"${_scopeId}>`);
            if (filters.status.length === 0) {
              _push2(`<!--[-->Status<!--]-->`);
            } else {
              _push2(`<!--[-->${ssrInterpolate(filters.status.length)} status<!--]-->`);
            }
            _push2(`</span>`);
            _push2(ssrRenderComponent(unref(ChevronDown), {
              class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
            }, null, _parent2, _scopeId));
            _push2(`</button>`);
          } else {
            return [
              createVNode("button", {
                onClick: toggle,
                class: "flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"
              }, [
                createVNode(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }),
                createVNode("span", { class: "text-sm text-slate-300" }, [
                  filters.status.length === 0 ? (openBlock(), createBlock(Fragment, { key: 0 }, [
                    createTextVNode("Status")
                  ], 64)) : (openBlock(), createBlock(Fragment, { key: 1 }, [
                    createTextVNode(toDisplayString(filters.status.length) + " status", 1)
                  ], 64))
                ]),
                createVNode(unref(ChevronDown), {
                  class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
                }, null, 8, ["class"])
              ], 8, ["onClick"])
            ];
          }
        }),
        default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<!--[-->`);
            ssrRenderList(statusOptions, (opt) => {
              _push2(`<button class="${ssrRenderClass([filters.status.includes(opt.value) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}><div class="${ssrRenderClass([filters.status.includes(opt.value) ? "bg-indigo-500 border-indigo-500" : "border-slate-600", "w-4 h-4 rounded border flex items-center justify-center transition-all"])}"${_scopeId}>`);
              if (filters.status.includes(opt.value)) {
                _push2(`<svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"${_scopeId}><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"${_scopeId}></path></svg>`);
              } else {
                _push2(`<!---->`);
              }
              _push2(`</div> ${ssrInterpolate(opt.label)}</button>`);
            });
            _push2(`<!--]-->`);
            if (filters.status.length > 0) {
              _push2(`<div class="border-t border-white/5 p-2"${_scopeId}><button class="flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(X), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              _push2(` Limpar filtros </button></div>`);
            } else {
              _push2(`<!---->`);
            }
          } else {
            return [
              (openBlock(), createBlock(Fragment, null, renderList(statusOptions, (opt) => {
                return createVNode("button", {
                  key: opt.value,
                  onClick: ($event) => toggleStatus(opt.value),
                  class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", filters.status.includes(opt.value) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
                }, [
                  createVNode("div", {
                    class: ["w-4 h-4 rounded border flex items-center justify-center transition-all", filters.status.includes(opt.value) ? "bg-indigo-500 border-indigo-500" : "border-slate-600"]
                  }, [
                    filters.status.includes(opt.value) ? (openBlock(), createBlock("svg", {
                      key: 0,
                      class: "w-3 h-3 text-white",
                      fill: "none",
                      viewBox: "0 0 24 24",
                      stroke: "currentColor"
                    }, [
                      createVNode("path", {
                        "stroke-linecap": "round",
                        "stroke-linejoin": "round",
                        "stroke-width": "3",
                        d: "M5 13l4 4L19 7"
                      })
                    ])) : createCommentVNode("", true)
                  ], 2),
                  createTextVNode(" " + toDisplayString(opt.label), 1)
                ], 10, ["onClick"]);
              }), 64)),
              filters.status.length > 0 ? (openBlock(), createBlock("div", {
                key: 0,
                class: "border-t border-white/5 p-2"
              }, [
                createVNode("button", {
                  onClick: ($event) => {
                    filters.status = [];
                    close();
                  },
                  class: "flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"
                }, [
                  createVNode(unref(X), { class: "w-4 h-4" }),
                  createTextVNode(" Limpar filtros ")
                ], 8, ["onClick"])
              ])) : createCommentVNode("", true)
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_router_link, {
        to: "/jobs/new",
        class: "btn-primary flex items-center gap-2"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Plus), { class: "w-4 h-4" }, null, _parent2, _scopeId));
            _push2(` Adicionar Vaga `);
          } else {
            return [
              createVNode(unref(Plus), { class: "w-4 h-4" }),
              createTextVNode(" Adicionar Vaga ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`<button class="btn-primary flex items-center gap-2">`);
      _push(ssrRenderComponent(unref(RefreshCcw), { class: "w-4 h-4" }, null, _parent));
      _push(` Sincronizar </button></div></div>`);
      if (unref(isLoading)) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (jobs.value.length === 0) {
        _push(`<div class="flex flex-col items-center justify-center h-64 text-center"><div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">`);
        _push(ssrRenderComponent(unref(X), { class: "w-10 h-10 text-slate-600" }, null, _parent));
        _push(`</div><p class="text-slate-500 max-w-md">Nenhum resultado</p></div>`);
      } else {
        _push(`<div class="glass-card overflow-hidden"><div class="overflow-x-auto"><table class="min-w-full divide-y divide-white/5"><thead class="bg-white/5"><tr><th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Título</th><th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Empresa</th><th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Status</th><th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Score</th><th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Coletado em</th><th class="px-8 py-5 text-right text-xs font-bold text-slate-400 uppercase tracking-widest">Ações</th></tr></thead><tbody class="divide-y divide-white/5"><!--[-->`);
        ssrRenderList(jobs.value, (job) => {
          var _a;
          _push(`<tr class="hover:bg-white/5 transition-colors group"><td class="px-8 py-6"><div class="flex flex-col"><div class="flex items-center gap-2"><span class="text-white font-bold group-hover:text-indigo-400 transition-colors">${ssrInterpolate(job.title)}</span><span class="text-[10px] text-slate-500 font-mono bg-slate-800/60 px-1.5 py-0.5 rounded">#${ssrInterpolate(job.id)}</span></div><span class="text-xs text-slate-500 font-mono mt-1">${ssrInterpolate(job.url.substring(0, 50))}...</span></div></td><td class="px-8 py-6"><div class="flex items-center gap-2"><div class="w-8 h-8 rounded-lg bg-slate-700/50 flex items-center justify-center text-xs font-bold text-slate-400 uppercase">${ssrInterpolate(((_a = job.company) == null ? void 0 : _a.charAt(0)) || "?")}</div><span class="text-slate-300">${ssrInterpolate(job.company || "Não identificada")}</span></div></td><td class="px-8 py-6"><span class="${ssrRenderClass([statusClass(job.status), "px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border"])}">${ssrInterpolate(statusLabelMap[job.status] || job.status)}</span></td><td class="px-8 py-6">`);
          if (job.max_score > 0) {
            _push(`<div class="${ssrRenderClass([scoreClass(job.max_score), "px-3 py-1 rounded-full text-xs font-bold text-center min-w-[3rem]"])}">${ssrInterpolate(Math.round(job.max_score))}% </div>`);
          } else {
            _push(`<span class="text-slate-600 text-sm">—</span>`);
          }
          _push(`</td><td class="px-8 py-6 text-slate-400 text-sm">${ssrInterpolate(formatDate(job.collected_at))}</td><td class="px-8 py-6 text-right"><div class="flex items-center justify-end gap-3">`);
          if (job.status !== "matched" && job.status !== "ignored") {
            _push(`<button class="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500 hover:text-white transition-all" title="Marcar como Match">`);
            _push(ssrRenderComponent(unref(CheckCircle2), { class: "w-4 h-4" }, null, _parent));
            _push(`</button>`);
          } else {
            _push(`<!---->`);
          }
          _push(`<a${ssrRenderAttr("href", job.url)} target="_blank" class="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg hover:bg-indigo-500 hover:text-white transition-all">`);
          _push(ssrRenderComponent(unref(ExternalLink), { class: "w-4 h-4" }, null, _parent));
          _push(`</a><button class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-red-500/20 hover:text-red-400 transition-all">`);
          _push(ssrRenderComponent(unref(Trash2), { class: "w-4 h-4" }, null, _parent));
          _push(`</button></div></td></tr>`);
        });
        _push(`<!--]--></tbody></table></div>`);
        if (totalPages.value > 0) {
          _push(`<div class="flex items-center justify-between px-8 py-4 border-t border-white/5"><span class="text-sm text-slate-400">${ssrInterpolate(showingInfo.value)}</span><div class="flex items-center gap-2"><button${ssrIncludeBooleanAttr(filters.page <= 1) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all">`);
          _push(ssrRenderComponent(unref(ChevronLeft), { class: "w-4 h-4 inline-block -ml-0.5" }, null, _parent));
          _push(` Anterior </button><span class="text-sm text-slate-400 px-2 font-mono">${ssrInterpolate(filters.page)} / ${ssrInterpolate(totalPages.value)}</span><button${ssrIncludeBooleanAttr(filters.page >= totalPages.value) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all"> Próximo `);
          _push(ssrRenderComponent(unref(ChevronRight), { class: "w-4 h-4 inline-block -mr-0.5" }, null, _parent));
          _push(`</button></div></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$w = _sfc_main$w.setup;
_sfc_main$w.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Jobs.vue");
  return _sfc_setup$w ? _sfc_setup$w(props, ctx) : void 0;
};
const _sfc_main$v = {
  __name: "JobAdd",
  __ssrInlineRender: true,
  setup(__props) {
    useRouter();
    const url = ref("");
    const extracting = ref(false);
    const saving = ref(false);
    const error = ref("");
    const extracted = ref(false);
    const sites = ref([]);
    const siteOpen = ref(false);
    const selectedSite = computed(() => {
      return sites.value.find((s) => s.id === form.site_id) || null;
    });
    const form = reactive({
      title: "",
      company: "",
      url: "",
      description: "",
      site_id: null
    });
    onMounted(async () => {
      try {
        const data = await listSites();
        sites.value = data || [];
      } catch {
        sites.value = [];
      }
    });
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="flex items-center justify-between"><div class="flex items-center gap-4">`);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/jobs",
        class: "p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-slate-600/50 transition-all"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(ArrowLeft), { class: "w-5 h-5" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(ArrowLeft), { class: "w-5 h-5" })
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`<div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Adicionar Vaga</h1><p class="text-slate-400">Cole o link da vaga para extrair os dados automaticamente.</p></div></div></div><div class="glass-card p-10"><form class="space-y-6"><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Fonte</label>`);
      _push(ssrRenderComponent(DropdownMenu, {
        open: siteOpen.value,
        "onUpdate:open": ($event) => siteOpen.value = $event,
        width: "w-64"
      }, {
        trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
          var _a, _b;
          if (_push2) {
            _push2(`<button type="button" class="flex items-center gap-2 px-4 py-2.5 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all w-full"${_scopeId}><span class="${ssrRenderClass([selectedSite.value ? "text-slate-300" : "text-slate-500", "text-sm flex-1 text-left"])}"${_scopeId}>${ssrInterpolate(((_a = selectedSite.value) == null ? void 0 : _a.name) || "Nenhuma")}</span>`);
            _push2(ssrRenderComponent(unref(ChevronDown), {
              class: ["w-3.5 h-3.5 text-slate-500 shrink-0", open ? "rotate-180" : ""]
            }, null, _parent2, _scopeId));
            _push2(`</button>`);
          } else {
            return [
              createVNode("button", {
                onClick: toggle,
                type: "button",
                class: "flex items-center gap-2 px-4 py-2.5 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all w-full"
              }, [
                createVNode("span", {
                  class: ["text-sm flex-1 text-left", selectedSite.value ? "text-slate-300" : "text-slate-500"]
                }, toDisplayString(((_b = selectedSite.value) == null ? void 0 : _b.name) || "Nenhuma"), 3),
                createVNode(unref(ChevronDown), {
                  class: ["w-3.5 h-3.5 text-slate-500 shrink-0", open ? "rotate-180" : ""]
                }, null, 8, ["class"])
              ], 8, ["onClick"])
            ];
          }
        }),
        default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button type="button" class="${ssrRenderClass([form.site_id === null ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}><div class="${ssrRenderClass([form.site_id === null ? "bg-indigo-500 border-indigo-500" : "border-slate-600", "w-4 h-4 rounded-full border flex items-center justify-center transition-all"])}"${_scopeId}>`);
            if (form.site_id === null) {
              _push2(`<div class="w-1.5 h-1.5 rounded-full bg-white"${_scopeId}></div>`);
            } else {
              _push2(`<!---->`);
            }
            _push2(`</div> Nenhuma </button><!--[-->`);
            ssrRenderList(sites.value, (site) => {
              _push2(`<button type="button" class="${ssrRenderClass([form.site_id === site.id ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}><div class="${ssrRenderClass([form.site_id === site.id ? "bg-indigo-500 border-indigo-500" : "border-slate-600", "w-4 h-4 rounded-full border flex items-center justify-center transition-all"])}"${_scopeId}>`);
              if (form.site_id === site.id) {
                _push2(`<div class="w-1.5 h-1.5 rounded-full bg-white"${_scopeId}></div>`);
              } else {
                _push2(`<!---->`);
              }
              _push2(`</div> ${ssrInterpolate(site.name)}</button>`);
            });
            _push2(`<!--]-->`);
          } else {
            return [
              createVNode("button", {
                onClick: ($event) => {
                  form.site_id = null;
                  close();
                },
                type: "button",
                class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", form.site_id === null ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
              }, [
                createVNode("div", {
                  class: ["w-4 h-4 rounded-full border flex items-center justify-center transition-all", form.site_id === null ? "bg-indigo-500 border-indigo-500" : "border-slate-600"]
                }, [
                  form.site_id === null ? (openBlock(), createBlock("div", {
                    key: 0,
                    class: "w-1.5 h-1.5 rounded-full bg-white"
                  })) : createCommentVNode("", true)
                ], 2),
                createTextVNode(" Nenhuma ")
              ], 10, ["onClick"]),
              (openBlock(true), createBlock(Fragment, null, renderList(sites.value, (site) => {
                return openBlock(), createBlock("button", {
                  key: site.id,
                  onClick: ($event) => {
                    form.site_id = site.id;
                    close();
                  },
                  type: "button",
                  class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", form.site_id === site.id ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
                }, [
                  createVNode("div", {
                    class: ["w-4 h-4 rounded-full border flex items-center justify-center transition-all", form.site_id === site.id ? "bg-indigo-500 border-indigo-500" : "border-slate-600"]
                  }, [
                    form.site_id === site.id ? (openBlock(), createBlock("div", {
                      key: 0,
                      class: "w-1.5 h-1.5 rounded-full bg-white"
                    })) : createCommentVNode("", true)
                  ], 2),
                  createTextVNode(" " + toDisplayString(site.name), 1)
                ], 10, ["onClick"]);
              }), 128))
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`</div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL da Vaga</label><div class="flex gap-3"><input${ssrRenderAttr("value", url.value)} type="url" class="input-field flex-1 h-12" placeholder="https://..." required><button type="submit" class="btn-primary h-12 px-8 flex items-center gap-2 whitespace-nowrap"${ssrIncludeBooleanAttr(extracting.value) ? " disabled" : ""}>`);
      if (extracting.value) {
        _push(ssrRenderComponent(unref(Loader2), { class: "w-5 h-5 animate-spin" }, null, _parent));
      } else {
        _push(ssrRenderComponent(unref(Search), { class: "w-5 h-5" }, null, _parent));
      }
      _push(` ${ssrInterpolate(extracting.value ? "Extraindo..." : "Extrair")}</button></div></div></form>`);
      if (error.value) {
        _push(`<div class="mt-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm">${ssrInterpolate(error.value)}</div>`);
      } else {
        _push(`<!---->`);
      }
      if (extracted.value) {
        _push(`<div class="mt-10 space-y-8"><div class="border-t border-white/5 pt-8"><h2 class="text-xl font-bold text-white font-outfit mb-6">Dados Extraídos</h2><div class="space-y-6"><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Título</label><input${ssrRenderAttr("value", form.title)} type="text" class="input-field w-full h-12" required></div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Empresa</label><input${ssrRenderAttr("value", form.company)} type="text" class="input-field w-full h-12"></div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL</label><input${ssrRenderAttr("value", form.url)} type="url" class="input-field w-full h-12 bg-slate-800/50 text-slate-500 cursor-not-allowed" readonly></div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Descrição</label><textarea class="input-field w-full h-48 resize-y">${ssrInterpolate(form.description)}</textarea></div></div><div class="flex gap-3 mt-8"><button class="btn-primary h-12 px-8 flex items-center gap-2"${ssrIncludeBooleanAttr(saving.value) ? " disabled" : ""}>`);
        if (saving.value) {
          _push(ssrRenderComponent(unref(Loader2), { class: "w-5 h-5 animate-spin" }, null, _parent));
        } else {
          _push(ssrRenderComponent(unref(CheckCircle2), { class: "w-5 h-5" }, null, _parent));
        }
        _push(` ${ssrInterpolate(saving.value ? "Salvando..." : "Salvar Vaga")}</button>`);
        _push(ssrRenderComponent(_component_router_link, {
          to: "/jobs",
          class: "h-12 px-8 flex items-center gap-2 bg-slate-700/50 text-slate-300 rounded-xl hover:bg-slate-600/50 transition-all"
        }, {
          default: withCtx((_, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(` Cancelar `);
            } else {
              return [
                createTextVNode(" Cancelar ")
              ];
            }
          }),
          _: 1
        }, _parent));
        _push(`</div></div></div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div></div>`);
    };
  }
};
const _sfc_setup$v = _sfc_main$v.setup;
_sfc_main$v.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/JobAdd.vue");
  return _sfc_setup$v ? _sfc_setup$v(props, ctx) : void 0;
};
const _sfc_main$u = {
  __name: "Matches",
  __ssrInlineRender: true,
  setup(__props) {
    const filters = reactive({
      sort: "desc",
      site: [],
      threshold: 65,
      applied: "false",
      page: 1,
      limit: 20
    });
    watch([() => filters.site, () => filters.sort, () => filters.threshold], () => {
      filters.page = 1;
    });
    const siteOpen = ref(false);
    const sortOpen = ref(false);
    const isRematching = ref(false);
    const rematchError = ref("");
    const sortLabel = computed(() => {
      return filters.sort === "desc" ? "Maior similaridade" : "Menor similaridade";
    });
    const toggleSite = (id) => {
      const idx = filters.site.indexOf(id);
      if (idx >= 0) {
        filters.site.splice(idx, 1);
      } else {
        filters.site.push(id);
      }
    };
    useQueryClient();
    const { data: matchesResponse, isLoading } = useMatches(filters);
    const { data: sites } = useSites();
    const matches = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.data) || [];
    });
    const total = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.total) || 0;
    });
    const totalPages = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.totalPages) || 0;
    });
    const showingInfo = computed(() => {
      if (total.value === 0) return "Nenhum match encontrado";
      const from = (filters.page - 1) * filters.limit + 1;
      const to = Math.min(filters.page * filters.limit, total.value);
      return `Mostrando ${from} a ${to} de ${total.value} matches`;
    });
    const scoreColor = (score) => {
      if (score >= 80) return "border-emerald-500";
      if (score >= 60) return "border-indigo-500";
      return "border-slate-700";
    };
    const scoreTextColor = (score) => {
      if (score >= 80) return "text-emerald-400";
      if (score >= 60) return "text-indigo-400";
      return "text-slate-400";
    };
    const formatDate = (date) => {
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "short"
      }).format(new Date(date));
    };
    const parseKeywords = (kws) => {
      if (!kws) return [];
      try {
        const parsed = JSON.parse(kws);
        return Array.isArray(parsed) ? parsed : [];
      } catch (e) {
        return kws.split(",").map((s) => s.trim());
      }
    };
    const extractTechnologies = (description) => {
      if (!description) return [];
      try {
        const text = description.toLowerCase();
        const techs = [
          "python",
          "javascript",
          "typescript",
          "java",
          "go",
          "golang",
          "rust",
          "c++",
          "c#",
          "csharp",
          "ruby",
          "php",
          "swift",
          "kotlin",
          "scala",
          "sql",
          "r",
          "matlab",
          "perl",
          "shell",
          "bash",
          "react",
          "vue",
          "angular",
          "svelte",
          "next.js",
          "nuxt",
          "node",
          "node.js",
          "deno",
          "bun",
          "django",
          "flask",
          "fastapi",
          "spring",
          "spring boot",
          "express",
          "nestjs",
          "koa",
          "gin",
          "docker",
          "kubernetes",
          "k8s",
          "aws",
          "azure",
          "gcp",
          "google cloud",
          "terraform",
          "ansible",
          "linux",
          "unix",
          "mysql",
          "postgresql",
          "postgres",
          "mongodb",
          "redis",
          "elasticsearch",
          "git",
          "github",
          "gitlab",
          "bitbucket",
          "ci/cd",
          "jenkins",
          "github actions",
          "gitlab ci",
          "devops",
          "agile",
          "scrum",
          "kanban",
          "api",
          "rest",
          "graphql",
          "grpc",
          "microservices",
          "machine learning",
          "ml",
          "ai",
          "data science",
          "pandas",
          "numpy",
          "tensorflow",
          "pytorch",
          "html",
          "css",
          "sass",
          "tailwind",
          "bootstrap",
          "webpack",
          "vite",
          "esbuild",
          "jest",
          "vitest",
          "cypress",
          "playwright",
          "testing library"
        ];
        const found = [];
        for (const tech of techs) {
          const escaped = tech.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
          const pattern = new RegExp(`\\b${escaped}\\b`, "i");
          if (pattern.test(text) && !found.includes(tech)) {
            found.push(tech);
          }
        }
        return found.slice(0, 8);
      } catch (e) {
        return [];
      }
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="flex items-center justify-between"><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Matches</h1><p class="text-slate-400">Correspondências inteligentes baseadas no seu currículo.</p></div><div class="flex items-center gap-4"><button${ssrIncludeBooleanAttr(isRematching.value) ? " disabled" : ""} class="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-xl hover:bg-indigo-500 transition-all whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed">`);
      _push(ssrRenderComponent(unref(RefreshCw), {
        class: ["w-4 h-4", isRematching.value ? "animate-spin" : ""]
      }, null, _parent));
      _push(`<span class="text-sm">${ssrInterpolate(isRematching.value ? "Reanalisando..." : "Re-analisar")}</span></button>`);
      _push(ssrRenderComponent(DropdownMenu, {
        open: siteOpen.value,
        "onUpdate:open": ($event) => siteOpen.value = $event,
        position: "bottom-right"
      }, {
        trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"${_scopeId}>`);
            _push2(ssrRenderComponent(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }, null, _parent2, _scopeId));
            _push2(`<span class="text-sm text-slate-300"${_scopeId}>`);
            if (filters.site.length === 0) {
              _push2(`<!--[-->Todos os Sites<!--]-->`);
            } else {
              _push2(`<!--[-->${ssrInterpolate(filters.site.length)} site${ssrInterpolate(filters.site.length > 1 ? "s" : "")}<!--]-->`);
            }
            _push2(`</span>`);
            _push2(ssrRenderComponent(unref(ChevronDown), {
              class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
            }, null, _parent2, _scopeId));
            _push2(`</button>`);
          } else {
            return [
              createVNode("button", {
                onClick: toggle,
                class: "flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"
              }, [
                createVNode(unref(Filter), { class: "w-4 h-4 text-slate-500 shrink-0" }),
                createVNode("span", { class: "text-sm text-slate-300" }, [
                  filters.site.length === 0 ? (openBlock(), createBlock(Fragment, { key: 0 }, [
                    createTextVNode("Todos os Sites")
                  ], 64)) : (openBlock(), createBlock(Fragment, { key: 1 }, [
                    createTextVNode(toDisplayString(filters.site.length) + " site" + toDisplayString(filters.site.length > 1 ? "s" : ""), 1)
                  ], 64))
                ]),
                createVNode(unref(ChevronDown), {
                  class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
                }, null, 8, ["class"])
              ], 8, ["onClick"])
            ];
          }
        }),
        default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<!--[-->`);
            ssrRenderList(unref(sites), (site) => {
              _push2(`<button class="${ssrRenderClass([filters.site.includes(site.id) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}><div class="${ssrRenderClass([filters.site.includes(site.id) ? "bg-indigo-500 border-indigo-500" : "border-slate-600", "w-4 h-4 rounded border flex items-center justify-center transition-all"])}"${_scopeId}>`);
              if (filters.site.includes(site.id)) {
                _push2(`<svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"${_scopeId}><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"${_scopeId}></path></svg>`);
              } else {
                _push2(`<!---->`);
              }
              _push2(`</div> ${ssrInterpolate(site.name)}</button>`);
            });
            _push2(`<!--]-->`);
            if (filters.site.length > 0) {
              _push2(`<div class="border-t border-white/5 p-2"${_scopeId}><button class="flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(X), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              _push2(` Limpar filtros </button></div>`);
            } else {
              _push2(`<!---->`);
            }
          } else {
            return [
              (openBlock(true), createBlock(Fragment, null, renderList(unref(sites), (site) => {
                return openBlock(), createBlock("button", {
                  key: site.id,
                  onClick: ($event) => toggleSite(site.id),
                  class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", filters.site.includes(site.id) ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
                }, [
                  createVNode("div", {
                    class: ["w-4 h-4 rounded border flex items-center justify-center transition-all", filters.site.includes(site.id) ? "bg-indigo-500 border-indigo-500" : "border-slate-600"]
                  }, [
                    filters.site.includes(site.id) ? (openBlock(), createBlock("svg", {
                      key: 0,
                      class: "w-3 h-3 text-white",
                      fill: "none",
                      viewBox: "0 0 24 24",
                      stroke: "currentColor"
                    }, [
                      createVNode("path", {
                        "stroke-linecap": "round",
                        "stroke-linejoin": "round",
                        "stroke-width": "3",
                        d: "M5 13l4 4L19 7"
                      })
                    ])) : createCommentVNode("", true)
                  ], 2),
                  createTextVNode(" " + toDisplayString(site.name), 1)
                ], 10, ["onClick"]);
              }), 128)),
              filters.site.length > 0 ? (openBlock(), createBlock("div", {
                key: 0,
                class: "border-t border-white/5 p-2"
              }, [
                createVNode("button", {
                  onClick: ($event) => {
                    filters.site = [];
                    close();
                  },
                  class: "flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all"
                }, [
                  createVNode(unref(X), { class: "w-4 h-4" }),
                  createTextVNode(" Limpar filtros ")
                ], 8, ["onClick"])
              ])) : createCommentVNode("", true)
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(DropdownMenu, {
        open: sortOpen.value,
        "onUpdate:open": ($event) => sortOpen.value = $event
      }, {
        trigger: withCtx(({ open, toggle }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"${_scopeId}>`);
            _push2(ssrRenderComponent(unref(ArrowUpDown), { class: "w-4 h-4 text-slate-500 shrink-0" }, null, _parent2, _scopeId));
            _push2(`<span class="text-sm text-slate-300"${_scopeId}>${ssrInterpolate(sortLabel.value)}</span>`);
            _push2(ssrRenderComponent(unref(ChevronDown), {
              class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
            }, null, _parent2, _scopeId));
            _push2(`</button>`);
          } else {
            return [
              createVNode("button", {
                onClick: toggle,
                class: "flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap"
              }, [
                createVNode(unref(ArrowUpDown), { class: "w-4 h-4 text-slate-500 shrink-0" }),
                createVNode("span", { class: "text-sm text-slate-300" }, toDisplayString(sortLabel.value), 1),
                createVNode(unref(ChevronDown), {
                  class: ["w-3.5 h-3.5 text-slate-500", open ? "rotate-180" : ""]
                }, null, 8, ["class"])
              ], 8, ["onClick"])
            ];
          }
        }),
        default: withCtx(({ close }, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<button class="${ssrRenderClass([filters.sort === "desc" ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}> Maior similaridade </button><button class="${ssrRenderClass([filters.sort === "asc" ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50", "flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"])}"${_scopeId}> Menor similaridade </button>`);
          } else {
            return [
              createVNode("button", {
                onClick: ($event) => {
                  filters.sort = "desc";
                  close();
                },
                class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", filters.sort === "desc" ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
              }, " Maior similaridade ", 10, ["onClick"]),
              createVNode("button", {
                onClick: ($event) => {
                  filters.sort = "asc";
                  close();
                },
                class: ["flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all", filters.sort === "asc" ? "bg-indigo-500/20 text-indigo-300" : "text-slate-400 hover:text-slate-300 hover:bg-slate-800/50"]
              }, " Menor similaridade ", 10, ["onClick"])
            ];
          }
        }),
        _: 1
      }, _parent));
      if (rematchError.value) {
        _push(`<p class="text-xs text-red-400 max-w-[200px]">${ssrInterpolate(rematchError.value)}</p>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div></div>`);
      if (unref(isLoading)) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (matches.value.length === 0) {
        _push(`<div class="flex flex-col items-center justify-center h-64 text-center"><div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">`);
        _push(ssrRenderComponent(unref(X), { class: "w-10 h-10 text-slate-600" }, null, _parent));
        _push(`</div><p class="text-slate-500 max-w-md">Nenhum resultado</p></div>`);
      } else {
        _push(`<div class="space-y-6"><div class="grid grid-cols-1 gap-6"><!--[-->`);
        ssrRenderList(matches.value, (match) => {
          var _a, _b, _c, _d, _e, _f;
          _push(`<div class="${ssrRenderClass([scoreColor(match.similarity_score), "glass-card p-8 group transition-all duration-500 hover:bg-slate-800/60 border-l-4"])}"><div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6"><div class="flex-1"><div class="flex items-center gap-3 mb-2"><h3 class="text-xl font-bold text-white font-outfit">${ssrInterpolate((_a = match.job) == null ? void 0 : _a.title)}</h3><span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">${ssrInterpolate((_b = match.job) == null ? void 0 : _b.id)}</span></div><div class="flex items-center gap-4 text-slate-400 text-sm"><span class="flex items-center gap-1.5">`);
          _push(ssrRenderComponent(unref(Building2), { class: "w-4 h-4" }, null, _parent));
          _push(` ${ssrInterpolate((_c = match.job) == null ? void 0 : _c.company)}</span><span class="w-1 h-1 bg-slate-700 rounded-full"></span><span class="flex items-center gap-1.5">`);
          _push(ssrRenderComponent(unref(Calendar), { class: "w-4 h-4" }, null, _parent));
          _push(` ${ssrInterpolate(formatDate(match.analyzed_at))}</span></div></div><div class="flex items-center gap-8"><div class="text-right"><div class="${ssrRenderClass([scoreTextColor(match.similarity_score), "text-4xl font-bold font-outfit"])}">${ssrInterpolate(Math.round(match.similarity_score))}% </div><div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Similaridade</div></div><button class="w-12 h-12 rounded-2xl bg-slate-700 flex items-center justify-center text-slate-400 hover:bg-red-500/20 hover:text-red-400 transition-all">`);
          _push(ssrRenderComponent(unref(EyeOff), { class: "w-6 h-6" }, null, _parent));
          _push(`</button><button class="${ssrRenderClass([match.applied ? "bg-emerald-600 hover:bg-emerald-500 shadow-emerald-900/20" : "bg-slate-700 hover:bg-indigo-600", "w-12 h-12 rounded-2xl flex items-center justify-center text-white shadow-lg transition-all active:scale-95"])}">`);
          if (match.applied) {
            _push(ssrRenderComponent(unref(CheckCircle), { class: "w-6 h-6" }, null, _parent));
          } else {
            _push(ssrRenderComponent(unref(Paperclip), { class: "w-6 h-6" }, null, _parent));
          }
          _push(`</button><a${ssrRenderAttr("href", (_d = match.job) == null ? void 0 : _d.url)} target="_blank" class="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/20 hover:scale-110 transition-transform active:scale-95">`);
          _push(ssrRenderComponent(unref(ExternalLink), { class: "w-6 h-6" }, null, _parent));
          _push(`</a></div></div><div class="mt-8 space-y-6">`);
          if ((_e = match.job) == null ? void 0 : _e.description) {
            _push(`<div class="flex flex-wrap gap-2"><!--[-->`);
            ssrRenderList(extractTechnologies(match.job.description), (tech) => {
              _push(`<div class="px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-xs text-emerald-300 font-medium">${ssrInterpolate(tech)}</div>`);
            });
            _push(`<!--]--></div>`);
          } else {
            _push(`<!---->`);
          }
          if ((_f = match.job) == null ? void 0 : _f.description) {
            _push(`<div class="bg-slate-900/50 border border-white/5 rounded-2xl p-4"><div class="text-xs text-slate-500 font-bold uppercase tracking-widest mb-2">Descrição da Vaga</div><p class="text-slate-400 text-sm line-clamp-4">${ssrInterpolate(match.job.description)}</p></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`<div class="flex flex-wrap gap-2"><!--[-->`);
          ssrRenderList(parseKeywords(match.keywords_matched), (kw) => {
            _push(`<div class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">${ssrInterpolate(kw)}</div>`);
          });
          _push(`<!--]--></div>`);
          if (match.ai_reason) {
            _push(`<div class="bg-slate-900/80 border border-white/5 rounded-2xl p-6 relative overflow-hidden"><div class="absolute top-0 right-0 p-4 opacity-10">`);
            _push(ssrRenderComponent(unref(Sparkles), { class: "w-12 h-12 text-indigo-400" }, null, _parent));
            _push(`</div><div class="flex items-center gap-2 mb-3 text-indigo-400 font-bold text-xs uppercase tracking-widest">`);
            _push(ssrRenderComponent(unref(BrainCircuit), { class: "w-4 h-4" }, null, _parent));
            _push(` Análise de Inteligência Artificial </div><p class="text-slate-300 text-sm leading-relaxed relative z-10">${ssrInterpolate(match.ai_reason)}</p></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div></div>`);
        });
        _push(`<!--]--></div>`);
        if (totalPages.value > 0) {
          _push(`<div class="flex items-center justify-between px-8 py-4 border-t border-white/5"><span class="text-sm text-slate-400">${ssrInterpolate(showingInfo.value)}</span><div class="flex items-center gap-2"><button${ssrIncludeBooleanAttr(filters.page <= 1) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all">`);
          _push(ssrRenderComponent(unref(ChevronLeft), { class: "w-4 h-4 inline-block -ml-0.5" }, null, _parent));
          _push(` Anterior </button><span class="text-sm text-slate-400 px-2 font-mono">${ssrInterpolate(filters.page)} / ${ssrInterpolate(totalPages.value)}</span><button${ssrIncludeBooleanAttr(filters.page >= totalPages.value) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all"> Próximo `);
          _push(ssrRenderComponent(unref(ChevronRight), { class: "w-4 h-4 inline-block -mr-0.5" }, null, _parent));
          _push(`</button></div></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$u = _sfc_main$u.setup;
_sfc_main$u.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Matches.vue");
  return _sfc_setup$u ? _sfc_setup$u(props, ctx) : void 0;
};
const _sfc_main$t = {
  __name: "Applied",
  __ssrInlineRender: true,
  setup(__props) {
    const filters = reactive({
      applied: "true",
      page: 1,
      limit: 20
    });
    useQueryClient();
    const { data: matchesResponse, isLoading } = useMatches(filters);
    const appliedJobs = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.data) || [];
    });
    const total = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.total) || 0;
    });
    const totalPages = computed(() => {
      var _a;
      return ((_a = matchesResponse.value) == null ? void 0 : _a.totalPages) || 0;
    });
    const generatingMatchId = ref(null);
    const prepsByMatch = reactive({});
    const loadPreps = async () => {
      try {
        const res = await listInterviewPreps({ limit: 100 });
        for (const prep of res.data || []) {
          prepsByMatch[prep.match_id] = prep;
        }
      } catch (e) {
      }
    };
    loadPreps();
    const showingInfo = computed(() => {
      if (total.value === 0) return "Nenhuma candidatura encontrada";
      const from = (filters.page - 1) * filters.limit + 1;
      const to = Math.min(filters.page * filters.limit, total.value);
      return `Mostrando ${from} a ${to} de ${total.value} candidaturas`;
    });
    const scoreColor = (score) => {
      if (score >= 80) return "border-emerald-500";
      if (score >= 60) return "border-indigo-500";
      return "border-slate-700";
    };
    const scoreTextColor = (score) => {
      if (score >= 80) return "text-emerald-400";
      if (score >= 60) return "text-indigo-400";
      return "text-slate-400";
    };
    const formatDate = (date) => {
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "short"
      }).format(new Date(date));
    };
    const parseKeywords = (kws) => {
      if (!kws) return [];
      try {
        const parsed = JSON.parse(kws);
        return Array.isArray(parsed) ? parsed : [];
      } catch (e) {
        return kws.split(",").map((s) => s.trim());
      }
    };
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Candidatadas</h1><p class="text-slate-400">Vagas onde você já se candidatou.</p></div>`);
      if (unref(isLoading)) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (appliedJobs.value.length === 0) {
        _push(`<div class="flex flex-col items-center justify-center h-64 text-center"><div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">`);
        _push(ssrRenderComponent(unref(X), { class: "w-10 h-10 text-slate-600" }, null, _parent));
        _push(`</div><p class="text-slate-500 max-w-md">Nenhum resultado</p></div>`);
      } else {
        _push(`<div class="space-y-6"><div class="grid grid-cols-1 gap-6"><!--[-->`);
        ssrRenderList(appliedJobs.value, (match) => {
          var _a, _b, _c, _d, _e;
          _push(`<div class="${ssrRenderClass([scoreColor(match.similarity_score), "glass-card p-8 group transition-all duration-500 hover:bg-slate-800/60 border-l-4"])}"><div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6"><div class="flex-1"><div class="flex items-center gap-3 mb-2"><h3 class="text-xl font-bold text-white font-outfit">${ssrInterpolate((_a = match.job) == null ? void 0 : _a.title)}</h3><span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">${ssrInterpolate((_b = match.job) == null ? void 0 : _b.id)}</span></div><div class="flex items-center gap-4 text-slate-400 text-sm"><span class="flex items-center gap-1.5">`);
          _push(ssrRenderComponent(unref(Building2), { class: "w-4 h-4" }, null, _parent));
          _push(` ${ssrInterpolate((_c = match.job) == null ? void 0 : _c.company)}</span><span class="w-1 h-1 bg-slate-700 rounded-full"></span><span class="flex items-center gap-1.5">`);
          _push(ssrRenderComponent(unref(Calendar), { class: "w-4 h-4" }, null, _parent));
          _push(` ${ssrInterpolate(formatDate(match.analyzed_at))}</span></div></div><div class="flex items-center gap-8"><div class="text-right"><div class="${ssrRenderClass([scoreTextColor(match.similarity_score), "text-4xl font-bold font-outfit"])}">${ssrInterpolate(Math.round(match.similarity_score))}% </div><div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Similaridade</div></div><div class="flex flex-col items-end gap-2">`);
          if (prepsByMatch[match.id]) {
            _push(`<div class="text-right">`);
            _push(ssrRenderComponent(_component_router_link, {
              to: `/interview-prep/${prepsByMatch[match.id].id}`,
              class: "inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-sm font-semibold hover:bg-indigo-500/25 transition-all"
            }, {
              default: withCtx((_, _push2, _parent2, _scopeId) => {
                if (_push2) {
                  _push2(ssrRenderComponent(unref(GraduationCap), { class: "w-4 h-4" }, null, _parent2, _scopeId));
                  _push2(` Preparação pronta `);
                } else {
                  return [
                    createVNode(unref(GraduationCap), { class: "w-4 h-4" }),
                    createTextVNode(" Preparação pronta ")
                  ];
                }
              }),
              _: 2
            }, _parent));
            _push(`<div class="mt-1 text-[10px] text-slate-500 font-mono">${ssrInterpolate(prepsByMatch[match.id].progress.practiced + prepsByMatch[match.id].progress.mastered)}/${ssrInterpolate(prepsByMatch[match.id].progress.total)} avaliadas </div></div>`);
          } else if (generatingMatchId.value === match.id) {
            _push(`<button disabled class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-500/20 text-indigo-300 text-sm font-semibold cursor-wait"><span class="w-4 h-4 border-2 border-indigo-400/30 border-t-indigo-400 rounded-full animate-spin"></span> Gerando... </button>`);
          } else {
            _push(`<button class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold shadow-lg shadow-indigo-500/20 transition-all active:scale-95">`);
            _push(ssrRenderComponent(unref(GraduationCap), { class: "w-4 h-4" }, null, _parent));
            _push(` Preparar entrevista </button>`);
          }
          _push(`</div><button class="w-12 h-12 rounded-2xl bg-red-500/20 hover:bg-red-500 flex items-center justify-center text-red-400 hover:text-white shadow-lg transition-all active:scale-95">`);
          _push(ssrRenderComponent(unref(Trash2), { class: "w-6 h-6" }, null, _parent));
          _push(`</button><a${ssrRenderAttr("href", (_d = match.job) == null ? void 0 : _d.url)} target="_blank" class="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/20 hover:scale-110 transition-transform active:scale-95">`);
          _push(ssrRenderComponent(unref(ExternalLink), { class: "w-6 h-6" }, null, _parent));
          _push(`</a></div></div><div class="mt-8 space-y-6">`);
          if ((_e = match.job) == null ? void 0 : _e.description) {
            _push(`<div class="bg-slate-900/50 border border-white/5 rounded-2xl p-4"><div class="text-xs text-slate-500 font-bold uppercase tracking-widest mb-2">Descrição da Vaga</div><p class="text-slate-400 text-sm line-clamp-4">${ssrInterpolate(match.job.description)}</p></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`<div class="flex flex-wrap gap-2"><!--[-->`);
          ssrRenderList(parseKeywords(match.keywords_matched), (kw) => {
            _push(`<div class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">${ssrInterpolate(kw)}</div>`);
          });
          _push(`<!--]--></div>`);
          if (match.ai_reason) {
            _push(`<div class="bg-slate-900/80 border border-white/5 rounded-2xl p-6 relative overflow-hidden"><div class="absolute top-0 right-0 p-4 opacity-10">`);
            _push(ssrRenderComponent(unref(Sparkles), { class: "w-12 h-12 text-indigo-400" }, null, _parent));
            _push(`</div><div class="flex items-center gap-2 mb-3 text-indigo-400 font-bold text-xs uppercase tracking-widest">`);
            _push(ssrRenderComponent(unref(BrainCircuit), { class: "w-4 h-4" }, null, _parent));
            _push(` Análise de Inteligência Artificial </div><p class="text-slate-300 text-sm leading-relaxed relative z-10">${ssrInterpolate(match.ai_reason)}</p></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div></div>`);
        });
        _push(`<!--]--></div>`);
        if (totalPages.value > 0) {
          _push(`<div class="flex items-center justify-between px-8 py-4 border-t border-white/5"><span class="text-sm text-slate-400">${ssrInterpolate(showingInfo.value)}</span><div class="flex items-center gap-2"><button${ssrIncludeBooleanAttr(filters.page <= 1) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all">`);
          _push(ssrRenderComponent(unref(ChevronLeft), { class: "w-4 h-4 inline-block -ml-0.5" }, null, _parent));
          _push(` Anterior </button><span class="text-sm text-slate-400 px-2 font-mono">${ssrInterpolate(filters.page)} / ${ssrInterpolate(totalPages.value)}</span><button${ssrIncludeBooleanAttr(filters.page >= totalPages.value) ? " disabled" : ""} class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all"> Próximo `);
          _push(ssrRenderComponent(unref(ChevronRight), { class: "w-4 h-4 inline-block -mr-0.5" }, null, _parent));
          _push(`</button></div></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$t = _sfc_main$t.setup;
_sfc_main$t.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Applied.vue");
  return _sfc_setup$t ? _sfc_setup$t(props, ctx) : void 0;
};
const _sfc_main$s = {
  __name: "Analysis",
  __ssrInlineRender: true,
  setup(__props) {
    useQueryClient();
    const resumeFile = ref(null);
    const currentAnalysis = ref(null);
    const isAnalyzing = ref(false);
    const selectedId = ref(null);
    const deletingId = ref(null);
    const { data: analyses, isLoading: analysesLoading } = useResumeAnalyses();
    const formatDate = (dateStr) => {
      if (!dateStr) return "-";
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit"
      }).format(new Date(dateStr));
    };
    const formatFileSize = (bytes) => {
      if (bytes < 1024) return bytes + " B";
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
      return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    };
    return (_ctx, _push, _parent, _attrs) => {
      var _a, _b, _c, _d, _e, _f, _g;
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Análise de Currículo</h1><p class="text-slate-400">Faça upload do seu currículo e receba feedback de melhorias e sugestões.</p></div><div class="glass-card p-8">`);
      if (!resumeFile.value && !currentAnalysis.value) {
        _push(`<div class="border-2 border-dashed border-white/10 rounded-2xl p-12 text-center hover:border-indigo-500/50 transition-colors"><div class="flex flex-col items-center gap-4"><div class="w-16 h-16 bg-indigo-500/10 rounded-full flex items-center justify-center">`);
        _push(ssrRenderComponent(unref(Upload), { class: "w-8 h-8 text-indigo-400" }, null, _parent));
        _push(`</div><div><label class="cursor-pointer"><span class="text-indigo-400 font-medium hover:text-indigo-300">Clique para fazer upload</span><input type="file" accept=".pdf,.doc,.docx,.txt" class="hidden"></label><span class="text-slate-500"> ou arraste o arquivo</span></div><div class="text-xs text-slate-500">PDF, DOC, DOCX ou TXT (max 5MB)</div></div></div>`);
      } else {
        _push(`<!---->`);
      }
      if (resumeFile.value && !currentAnalysis.value) {
        _push(`<div class="space-y-4"><div class="flex items-center justify-between p-4 bg-slate-800/50 rounded-xl"><div class="flex items-center gap-3">`);
        _push(ssrRenderComponent(unref(FileText), { class: "w-8 h-8 text-indigo-400" }, null, _parent));
        _push(`<div><div class="text-white font-medium">${ssrInterpolate(resumeFile.value.name)}</div><div class="text-xs text-slate-500">${ssrInterpolate(formatFileSize(resumeFile.value.size))}</div></div></div><button${ssrIncludeBooleanAttr(isAnalyzing.value) ? " disabled" : ""} class="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-600 rounded-xl text-white font-medium transition-colors flex items-center gap-2">`);
        if (isAnalyzing.value) {
          _push(ssrRenderComponent(unref(Loader), { class: "w-5 h-5 animate-spin" }, null, _parent));
        } else {
          _push(`<!---->`);
        }
        _push(`<span>${ssrInterpolate(isAnalyzing.value ? "Analisando..." : "Analisar")}</span></button></div></div>`);
      } else {
        _push(`<!---->`);
      }
      if (isAnalyzing.value) {
        _push(`<div class="flex flex-col items-center justify-center py-12">`);
        _push(ssrRenderComponent(unref(Loader), { class: "w-12 h-12 text-indigo-400 animate-spin mb-4" }, null, _parent));
        _push(`<div class="text-slate-400">Analisando seu currículo...</div><div class="text-xs text-slate-500 mt-2">Isso pode levar alguns segundos</div></div>`);
      } else {
        _push(`<!---->`);
      }
      if (currentAnalysis.value) {
        _push(`<div class="space-y-6"><div class="flex items-center justify-between"><button class="text-slate-400 hover:text-white text-sm flex items-center gap-2">`);
        _push(ssrRenderComponent(unref(ArrowLeft), { class: "w-4 h-4" }, null, _parent));
        _push(`<span>Nova análise</span></button><div class="text-xs text-slate-500">${ssrInterpolate(currentAnalysis.value.file_name)}</div></div><div class="grid grid-cols-1 md:grid-cols-3 gap-4"><div class="bg-emerald-500/10 border border-emerald-500/20 rounded-2xl p-6"><div class="text-emerald-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos Fortes</div><div class="text-3xl font-bold text-white">${ssrInterpolate(((_a = currentAnalysis.value.strengths) == null ? void 0 : _a.length) || 0)}</div><div class="text-xs text-slate-500">itens identificados</div></div><div class="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-6"><div class="text-amber-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos de Atenção</div><div class="text-3xl font-bold text-white">${ssrInterpolate(((_b = currentAnalysis.value.weaknesses) == null ? void 0 : _b.length) || 0)}</div><div class="text-xs text-slate-500">itens identificados</div></div><div class="bg-indigo-500/10 border border-indigo-500/20 rounded-2xl p-6"><div class="text-indigo-400 font-bold text-xs uppercase tracking-widest mb-2">Sugestões</div><div class="text-3xl font-bold text-white">${ssrInterpolate(((_c = currentAnalysis.value.suggestions) == null ? void 0 : _c.length) || 0)}</div><div class="text-xs text-slate-500">recomendações</div></div></div>`);
        if ((_d = currentAnalysis.value.strengths) == null ? void 0 : _d.length) {
          _push(`<div class="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6"><div class="flex items-center gap-2 mb-4 text-emerald-400 font-bold text-sm uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(ThumbsUp), { class: "w-5 h-5" }, null, _parent));
          _push(` Pontos Fortes </div><ul class="space-y-3"><!--[-->`);
          ssrRenderList(currentAnalysis.value.strengths, (strength, idx) => {
            _push(`<li class="flex gap-3 text-slate-300">`);
            _push(ssrRenderComponent(unref(CheckCircle), { class: "w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" }, null, _parent));
            _push(`<span>${ssrInterpolate(strength)}</span></li>`);
          });
          _push(`<!--]--></ul></div>`);
        } else {
          _push(`<!---->`);
        }
        if ((_e = currentAnalysis.value.weaknesses) == null ? void 0 : _e.length) {
          _push(`<div class="bg-amber-500/5 border border-amber-500/20 rounded-2xl p-6"><div class="flex items-center gap-2 mb-4 text-amber-400 font-bold text-sm uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(AlertTriangle), { class: "w-5 h-5" }, null, _parent));
          _push(` Pontos de Atenção </div><ul class="space-y-3"><!--[-->`);
          ssrRenderList(currentAnalysis.value.weaknesses, (weakness, idx) => {
            _push(`<li class="flex gap-3 text-slate-300">`);
            _push(ssrRenderComponent(unref(AlertCircle), { class: "w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" }, null, _parent));
            _push(`<span>${ssrInterpolate(weakness)}</span></li>`);
          });
          _push(`<!--]--></ul></div>`);
        } else {
          _push(`<!---->`);
        }
        if ((_f = currentAnalysis.value.suggestions) == null ? void 0 : _f.length) {
          _push(`<div class="bg-indigo-500/5 border border-indigo-500/20 rounded-2xl p-6"><div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(Lightbulb), { class: "w-5 h-5" }, null, _parent));
          _push(` Sugestões de Melhoria </div><ul class="space-y-3"><!--[-->`);
          ssrRenderList(currentAnalysis.value.suggestions, (suggestion, idx) => {
            _push(`<li class="flex gap-3 text-slate-300">`);
            _push(ssrRenderComponent(unref(ChevronRight), { class: "w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" }, null, _parent));
            _push(`<span>${ssrInterpolate(suggestion)}</span></li>`);
          });
          _push(`<!--]--></ul></div>`);
        } else {
          _push(`<!---->`);
        }
        if (currentAnalysis.value.fullAnalysis) {
          _push(`<div class="bg-slate-900/50 border border-white/5 rounded-2xl p-6"><div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(BrainCircuit), { class: "w-5 h-5" }, null, _parent));
          _push(` Análise Completa </div><div class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">${ssrInterpolate(currentAnalysis.value.fullAnalysis)}</div></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div><div class="glass-card p-8"><div class="flex items-center gap-3 mb-8"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
      _push(ssrRenderComponent(unref(History), { class: "w-6 h-6" }, null, _parent));
      _push(`</div><h2 class="text-2xl font-bold text-white font-outfit">Histórico de Análises</h2></div>`);
      if (unref(analysesLoading)) {
        _push(`<div class="flex justify-center py-12"><div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (!((_g = unref(analyses)) == null ? void 0 : _g.length)) {
        _push(`<div class="text-center py-12 bg-slate-950/30 rounded-3xl border border-white/5">`);
        _push(ssrRenderComponent(unref(FileX), { class: "w-12 h-12 text-slate-500 mx-auto mb-4 opacity-20" }, null, _parent));
        _push(`<p class="text-slate-400 font-medium">Nenhuma análise encontrada.</p><p class="text-xs text-slate-500 mt-2">Faça o upload de um currículo acima para começar.</p></div>`);
      } else {
        _push(`<div class="space-y-3"><!--[-->`);
        ssrRenderList(unref(analyses), (item) => {
          var _a2, _b2, _c2;
          _push(`<div class="${ssrRenderClass([
            "flex items-center justify-between p-4 rounded-xl border transition-all group",
            selectedId.value === item.id ? "bg-indigo-500/10 border-indigo-500/30 cursor-default" : "bg-slate-950/50 border-white/5 cursor-pointer hover:border-indigo-500/20"
          ])}"><div class="flex items-center gap-4 min-w-0 flex-1"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400 flex-shrink-0">`);
          _push(ssrRenderComponent(unref(FileText), { class: "w-5 h-5" }, null, _parent));
          _push(`</div><div class="min-w-0"><div class="text-white font-medium truncate">${ssrInterpolate(item.file_name)}</div><div class="text-xs text-slate-500 mt-1">${ssrInterpolate(formatDate(item.created_at))}</div></div></div><div class="flex items-center gap-4 flex-shrink-0 ml-4"><div class="flex items-center gap-3 text-xs text-slate-500"><span class="text-emerald-400">${ssrInterpolate(((_a2 = item.strengths) == null ? void 0 : _a2.length) || 0)} fortes</span><span class="text-amber-400">${ssrInterpolate(((_b2 = item.weaknesses) == null ? void 0 : _b2.length) || 0)} atencao</span><span class="text-indigo-400">${ssrInterpolate(((_c2 = item.suggestions) == null ? void 0 : _c2.length) || 0)} sugestoes</span></div><button${ssrIncludeBooleanAttr(deletingId.value === item.id) ? " disabled" : ""} class="p-2 rounded-lg text-slate-500 hover:bg-red-500/20 hover:text-red-400 transition-all opacity-0 group-hover:opacity-100">`);
          if (deletingId.value !== item.id) {
            _push(ssrRenderComponent(unref(Trash2), { class: "w-4 h-4" }, null, _parent));
          } else {
            _push(`<div class="w-4 h-4 border-2 border-red-400/30 border-t-red-400 rounded-full animate-spin"></div>`);
          }
          _push(`</button></div></div>`);
        });
        _push(`<!--]--></div>`);
      }
      _push(`</div></div>`);
    };
  }
};
const _sfc_setup$s = _sfc_main$s.setup;
_sfc_main$s.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Analysis.vue");
  return _sfc_setup$s ? _sfc_setup$s(props, ctx) : void 0;
};
const _sfc_main$r = {
  __name: "Settings",
  __ssrInlineRender: true,
  setup(__props) {
    const queryClient = useQueryClient();
    const { updateUser, user } = useAuth();
    const isAdmin = computed(() => {
      var _a;
      return ((_a = user.value) == null ? void 0 : _a.role) === "admin";
    });
    const newSite = ref({ name: "", url: "" });
    const newKeyword = ref("");
    const keywords = ref([]);
    const sitesQuery = useSites({ enabled: !isAdmin.value });
    const meQuery = useMe();
    const negativeKeywordsQuery = useNegativeKeywords({ enabled: !isAdmin.value });
    const sites = sitesQuery.data;
    const sitesLoading = sitesQuery.isLoading;
    const keywordsLoading = negativeKeywordsQuery.isLoading;
    watch(() => {
      var _a;
      return (_a = negativeKeywordsQuery.data) == null ? void 0 : _a.value;
    }, (list) => {
      if (Array.isArray(list)) {
        keywords.value = [...list];
      }
    }, { immediate: true });
    const profileForm = ref({ name: "", city: "" });
    watch(() => {
      var _a;
      return (_a = meQuery.data) == null ? void 0 : _a.value;
    }, (resp) => {
      if (resp == null ? void 0 : resp.user) {
        profileForm.value = {
          name: resp.user.name || "",
          city: resp.user.city || ""
        };
      }
    }, { immediate: true });
    const profileMutation = useMutation({
      mutationFn: updateProfile,
      onSuccess: (response) => {
        var _a;
        queryClient.invalidateQueries({ queryKey: ["me"] });
        if ((_a = response == null ? void 0 : response.data) == null ? void 0 : _a.user) {
          updateUser(response.data.user);
        }
      }
    });
    const siteMutation = useMutation({
      mutationFn: addSite,
      onSuccess: () => {
        newSite.value = { name: "", url: "" };
        queryClient.invalidateQueries({ queryKey: ["sites"] });
      }
    });
    useMutation({
      mutationFn: deleteSite,
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["sites"] });
      }
    });
    useMutation({
      mutationFn: ({ id, data }) => updateSite(id, data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["sites"] });
      }
    });
    useMutation({
      mutationFn: updateNegativeKeywords,
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["negative-keywords"] });
      }
    });
    return (_ctx, _push, _parent, _attrs) => {
      var _a;
      const _component_CheckCircle2 = resolveComponent("CheckCircle2");
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-12" }, _attrs))}><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Configurações</h1><p class="text-slate-400">${ssrInterpolate(isAdmin.value ? "Atualize seus dados pessoais." : "Personalize seu scanner e gerencie seu plano.")}</p></div><div class="glass-card p-10"><div class="flex items-center gap-3 mb-8"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
      _push(ssrRenderComponent(unref(MapPin), { class: "w-6 h-6" }, null, _parent));
      _push(`</div><div><h2 class="text-2xl font-bold text-white font-outfit">Seu Perfil</h2><p class="text-sm text-slate-400">Configure sua localização para melhor matching</p></div></div><form class="space-y-6"><div class="grid grid-cols-1 md:grid-cols-2 gap-6"><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Nome</label><input${ssrRenderAttr("value", profileForm.value.name)} type="text" class="input-field w-full h-12" placeholder="Seu nome"></div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Cidade</label><input${ssrRenderAttr("value", profileForm.value.city)} type="text" class="input-field w-full h-12" placeholder="ex: São Paulo, Rio de Janeiro"><p class="text-xs text-slate-500 ml-1">Usada para filtrar vagas presenciais e híbridas</p></div></div><button type="submit" class="h-10 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2"${ssrIncludeBooleanAttr(unref(profileMutation).isPending.value) ? " disabled" : ""}>`);
      if (!unref(profileMutation).isPending.value) {
        _push(ssrRenderComponent(_component_CheckCircle2, { class: "w-4 h-4" }, null, _parent));
      } else {
        _push(`<div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>`);
      }
      _push(` ${ssrInterpolate(unref(profileMutation).isPending.value ? "Salvando..." : "Salvar Perfil")}</button></form></div>`);
      if (!isAdmin.value) {
        _push(`<div class="grid grid-cols-1 gap-6">`);
        _push(ssrRenderComponent(_component_router_link, {
          to: "/settings/billing",
          class: "glass-card p-6 hover:border-indigo-500/30 transition-all cursor-pointer group"
        }, {
          default: withCtx((_, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(`<div class="flex items-center gap-4"${_scopeId}><div class="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400 group-hover:bg-emerald-500/20 transition-colors"${_scopeId}>`);
              _push2(ssrRenderComponent(unref(CreditCard), { class: "w-6 h-6" }, null, _parent2, _scopeId));
              _push2(`</div><div${_scopeId}><h3 class="font-bold text-white"${_scopeId}>Billing</h3><p class="text-sm text-slate-400"${_scopeId}>Plano, pagamentos e limites</p></div></div>`);
            } else {
              return [
                createVNode("div", { class: "flex items-center gap-4" }, [
                  createVNode("div", { class: "w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400 group-hover:bg-emerald-500/20 transition-colors" }, [
                    createVNode(unref(CreditCard), { class: "w-6 h-6" })
                  ]),
                  createVNode("div", null, [
                    createVNode("h3", { class: "font-bold text-white" }, "Billing"),
                    createVNode("p", { class: "text-sm text-slate-400" }, "Plano, pagamentos e limites")
                  ])
                ])
              ];
            }
          }),
          _: 1
        }, _parent));
        _push(`</div>`);
      } else {
        _push(`<!---->`);
      }
      if (!isAdmin.value) {
        _push(`<div class="grid grid-cols-1 lg:grid-cols-2 gap-12"><div class="glass-card p-10"><div class="flex items-center gap-3 mb-8"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
        _push(ssrRenderComponent(unref(Globe), { class: "w-6 h-6" }, null, _parent));
        _push(`</div><h2 class="text-2xl font-bold text-white font-outfit">Adicionar Fonte</h2></div><form class="space-y-6"><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Nome da Plataforma</label><input${ssrRenderAttr("value", newSite.value.name)} type="text" class="input-field w-full h-12" placeholder="ex: LinkedIn, RemoteOK" required></div><div class="space-y-2"><label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL de Busca</label><input${ssrRenderAttr("value", newSite.value.url)} type="url" class="input-field w-full h-12" placeholder="https://..." required></div><button type="submit" class="btn-primary w-full h-12 flex items-center justify-center gap-2"${ssrIncludeBooleanAttr(unref(siteMutation).isPending.value) ? " disabled" : ""}>`);
        if (!unref(siteMutation).isPending.value) {
          _push(ssrRenderComponent(unref(Plus), { class: "w-5 h-5" }, null, _parent));
        } else {
          _push(`<div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>`);
        }
        _push(` ${ssrInterpolate(unref(siteMutation).isPending.value ? "Analisando site com IA..." : "Adicionar Fonte")}</button></form></div><div class="glass-card p-10"><div class="flex items-center justify-between mb-10"><h2 class="text-2xl font-bold text-white font-outfit">Fontes Cadastradas</h2>`);
        _push(ssrRenderComponent(unref(Globe), { class: "text-slate-500 w-6 h-6" }, null, _parent));
        _push(`</div>`);
        if (unref(sitesLoading)) {
          _push(`<div class="flex justify-center py-12"><div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
        } else if (!((_a = unref(sites)) == null ? void 0 : _a.length)) {
          _push(`<div class="text-center py-20 bg-slate-950/30 rounded-3xl border border-white/5"><div class="text-slate-500 mb-4 flex justify-center">`);
          _push(ssrRenderComponent(unref(Globe), { class: "w-12 h-12 opacity-20" }, null, _parent));
          _push(`</div><p class="text-slate-400 font-medium">Nenhuma fonte cadastrada.</p></div>`);
        } else {
          _push(`<div class="space-y-4"><!--[-->`);
          ssrRenderList(unref(sites), (site) => {
            var _a2;
            _push(`<div class="bg-slate-950/50 border border-white/5 rounded-2xl p-5 flex items-center justify-between hover:border-indigo-500/30 transition-all"><div class="flex items-center gap-4"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
            _push(ssrRenderComponent(unref(Globe), { class: "w-5 h-5" }, null, _parent));
            _push(`</div><div><h3 class="font-bold text-white">${ssrInterpolate(site.name)}</h3><p class="text-xs text-slate-500 mt-1 flex items-center gap-2"><a${ssrRenderAttr("href", site.url)} target="_blank" class="hover:text-indigo-400 transition-colors flex items-center gap-1">${ssrInterpolate((_a2 = site.url) == null ? void 0 : _a2.substring(0, 40))}... `);
            _push(ssrRenderComponent(unref(ExternalLink), { class: "w-3 h-3" }, null, _parent));
            _push(`</a></p></div></div><div class="flex items-center gap-3"><button class="${ssrRenderClass([site.active ? "bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20" : "bg-slate-700/50 text-slate-400 hover:bg-slate-600/50", "p-2 rounded-lg transition-all"])}"${ssrRenderAttr("title", site.active ? "Desativar" : "Ativar")}>`);
            _push(ssrRenderComponent(unref(Power), { class: "w-4 h-4" }, null, _parent));
            _push(`</button><button class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-red-500/20 hover:text-red-400 transition-all">`);
            _push(ssrRenderComponent(unref(Trash2), { class: "w-4 h-4" }, null, _parent));
            _push(`</button></div></div>`);
          });
          _push(`<!--]--></div>`);
        }
        _push(`</div></div>`);
      } else {
        _push(`<!---->`);
      }
      if (!isAdmin.value) {
        _push(`<div class="glass-card p-10"><div class="flex items-center gap-3 mb-8"><div class="w-10 h-10 bg-red-500/10 rounded-xl flex items-center justify-center text-red-400">`);
        _push(ssrRenderComponent(unref(Ban), { class: "w-6 h-6" }, null, _parent));
        _push(`</div><div><h2 class="text-2xl font-bold text-white font-outfit">Palavras-chave de bloqueio</h2><p class="text-sm text-slate-400">Vagas que citarem essas palavras perdem 3 pontos no score de match.</p></div></div><div class="space-y-6"><div class="flex gap-3"><input${ssrRenderAttr("value", newKeyword.value)} type="text" class="input-field flex-1 h-12" placeholder="ex: freela, CLT, bilingue..."><button type="button" class="h-12 px-5 bg-red-600/80 hover:bg-red-500 text-white font-bold rounded-xl transition-all flex items-center gap-2">`);
        _push(ssrRenderComponent(unref(Plus), { class: "w-5 h-5" }, null, _parent));
        _push(` Adicionar </button></div>`);
        if (unref(keywordsLoading)) {
          _push(`<div class="flex justify-center py-8"><div class="w-8 h-8 border-4 border-red-500/30 border-t-red-500 rounded-full animate-spin"></div></div>`);
        } else if (!keywords.value.length) {
          _push(`<div class="text-center py-10 bg-slate-950/30 rounded-2xl border border-white/5"><p class="text-slate-400 font-medium">Nenhuma palavra de bloqueio configurada.</p></div>`);
        } else {
          _push(`<div class="flex flex-wrap gap-3"><!--[-->`);
          ssrRenderList(keywords.value, (kw, i) => {
            _push(`<span class="flex items-center gap-2 px-4 py-2 bg-red-500/10 border border-red-500/20 rounded-xl text-sm text-red-300 font-medium">${ssrInterpolate(kw)} <button class="text-red-400/60 hover:text-red-300 transition-colors"${ssrRenderAttr("title", `Remover ${kw}`)}>`);
            _push(ssrRenderComponent(unref(X), { class: "w-4 h-4" }, null, _parent));
            _push(`</button></span>`);
          });
          _push(`<!--]--></div>`);
        }
        _push(`</div></div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$r = _sfc_main$r.setup;
_sfc_main$r.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Settings.vue");
  return _sfc_setup$r ? _sfc_setup$r(props, ctx) : void 0;
};
const _sfc_main$q = {
  __name: "Billing",
  __ssrInlineRender: true,
  setup(__props) {
    const { data: meData, isLoading: meLoading, isError: meError } = useMe();
    const { data: plansData, isLoading: plansLoading, isError: plansError } = usePlans();
    const planName = computed(() => {
      var _a, _b;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.name) || "-";
    });
    const planSlug = computed(() => {
      var _a, _b, _c, _d;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.slug) || ((_d = (_c = meData.value) == null ? void 0 : _c.user) == null ? void 0 : _d.plan) || "free";
    });
    const planPrice = computed(() => {
      var _a, _b;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.price_monthly) ?? 0;
    });
    const maxJobsRaw = computed(() => {
      var _a, _b;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.max_jobs) ?? -1;
    });
    const maxResumesRaw = computed(() => {
      var _a, _b;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.max_resumes) ?? -1;
    });
    const maxSitesRaw = computed(() => {
      var _a, _b;
      return ((_b = (_a = meData.value) == null ? void 0 : _a.plan) == null ? void 0 : _b.max_sites) ?? -1;
    });
    const maxJobs = computed(() => maxJobsRaw.value === -1 ? "∞" : maxJobsRaw.value);
    const maxResumes = computed(() => maxResumesRaw.value === -1 ? "∞" : maxResumesRaw.value);
    const maxSites = computed(() => maxSitesRaw.value === -1 ? "∞" : maxSitesRaw.value);
    const usage = computed(() => {
      var _a;
      return ((_a = meData.value) == null ? void 0 : _a.usage) || { jobs: 0, resumes: 0, sites: 0 };
    });
    const changingPlan = ref(false);
    function usageBar(current, max) {
      if (max === -1) return 0;
      return Math.min(100, current / max * 100);
    }
    function parseFeatures(features) {
      if (!features) return [];
      try {
        const parsed = JSON.parse(features);
        return Array.isArray(parsed) ? parsed : [];
      } catch {
        return [];
      }
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="mb-8"><h1 class="text-2xl font-bold font-outfit text-white">Billing</h1><p class="text-slate-400 mt-1">Gerencie seu plano e pagamentos</p></div>`);
      if (unref(meLoading)) {
        _push(`<div class="text-slate-400 text-center py-12">Carregando...</div>`);
      } else if (unref(meError)) {
        _push(`<div class="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm"> Erro ao carregar dados do plano. </div>`);
      } else if (unref(meData)) {
        _push(`<!--[--><div class="glass-card p-6 mb-8"><div class="flex items-center justify-between"><div><div class="text-sm text-slate-400">Plano atual</div><div class="text-2xl font-bold text-white mt-1">${ssrInterpolate(planName.value)}</div><div class="text-sm text-slate-400 mt-1">${ssrInterpolate(planPrice.value === 0 ? "Grátis" : `R$ ${planPrice.value / 100}/mês`)}</div></div>`);
        if (planSlug.value === "free") {
          _push(`<button${ssrIncludeBooleanAttr(changingPlan.value) ? " disabled" : ""} class="btn-primary flex items-center gap-2">`);
          if (changingPlan.value) {
            _push(ssrRenderComponent(unref(Loader2), { class: "w-4 h-4 animate-spin" }, null, _parent));
          } else {
            _push(`<!---->`);
          }
          _push(` Upgrade para Pro </button>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div><div class="mt-6 pt-6 border-t border-white/5 grid grid-cols-3 gap-6"><div><div class="text-xs text-slate-500">Vagas</div><div class="text-lg font-bold text-white mt-1">${ssrInterpolate(usage.value.jobs)} / ${ssrInterpolate(maxJobs.value)}</div><div class="w-full bg-slate-700 rounded-full h-2 mt-2"><div class="bg-indigo-500 h-2 rounded-full" style="${ssrRenderStyle({ width: usageBar(usage.value.jobs, maxJobsRaw.value) + "%" })}"></div></div></div><div><div class="text-xs text-slate-500">Currículos</div><div class="text-lg font-bold text-white mt-1">${ssrInterpolate(usage.value.resumes)} / ${ssrInterpolate(maxResumes.value)}</div><div class="w-full bg-slate-700 rounded-full h-2 mt-2"><div class="bg-indigo-500 h-2 rounded-full" style="${ssrRenderStyle({ width: usageBar(usage.value.resumes, maxResumesRaw.value) + "%" })}"></div></div></div><div><div class="text-xs text-slate-500">Sites</div><div class="text-lg font-bold text-white mt-1">${ssrInterpolate(usage.value.sites)} / ${ssrInterpolate(maxSites.value)}</div><div class="w-full bg-slate-700 rounded-full h-2 mt-2"><div class="bg-indigo-500 h-2 rounded-full" style="${ssrRenderStyle({ width: usageBar(usage.value.sites, maxSitesRaw.value) + "%" })}"></div></div></div></div></div><h2 class="text-lg font-semibold text-white mb-4">Planos disponíveis</h2>`);
        if (unref(plansLoading)) {
          _push(`<div class="text-slate-500 text-sm">Carregando planos...</div>`);
        } else if (unref(plansError)) {
          _push(`<div class="text-slate-500 text-sm">Não foi possível carregar planos.</div>`);
        } else {
          _push(`<div class="grid grid-cols-2 gap-6"><!--[-->`);
          ssrRenderList(unref(plansData), (plan) => {
            _push(`<div class="${ssrRenderClass(["glass-card p-6", planSlug.value === plan.slug ? "ring-2 ring-indigo-500" : ""])}"><div class="text-sm text-slate-400 mb-2">${ssrInterpolate(plan.name)}</div><div class="text-3xl font-bold text-white">${ssrInterpolate(plan.price_monthly === 0 ? "Grátis" : `R$ ${plan.price_monthly / 100}`)} `);
            if (plan.price_monthly > 0) {
              _push(`<span class="text-sm font-normal text-slate-400">/mês</span>`);
            } else {
              _push(`<!---->`);
            }
            _push(`</div><ul class="mt-4 space-y-2"><!--[-->`);
            ssrRenderList(parseFeatures(plan.features), (feature) => {
              _push(`<li class="text-sm text-slate-300 flex items-center gap-2">`);
              _push(ssrRenderComponent(unref(Check), { class: "w-4 h-4 text-emerald-400" }, null, _parent));
              _push(` ${ssrInterpolate(feature)}</li>`);
            });
            _push(`<!--]--></ul>`);
            if (planSlug.value !== plan.slug) {
              _push(`<button${ssrIncludeBooleanAttr(changingPlan.value) ? " disabled" : ""} class="w-full mt-6 py-3 bg-slate-700 hover:bg-slate-600 disabled:opacity-50 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2">`);
              if (changingPlan.value) {
                _push(ssrRenderComponent(unref(Loader2), { class: "w-4 h-4 animate-spin" }, null, _parent));
              } else {
                _push(`<!---->`);
              }
              _push(` ${ssrInterpolate(planSlug.value === "free" ? "Começar" : "Downgrade")}</button>`);
            } else {
              _push(`<button class="w-full mt-6 py-3 bg-indigo-600 text-white font-medium rounded-lg cursor-default"> Plano atual </button>`);
            }
            _push(`</div>`);
          });
          _push(`<!--]--></div>`);
        }
        _push(`<!--]-->`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$q = _sfc_main$q.setup;
_sfc_main$q.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Billing.vue");
  return _sfc_setup$q ? _sfc_setup$q(props, ctx) : void 0;
};
const _sfc_main$p = {
  __name: "Login",
  __ssrInlineRender: true,
  setup(__props) {
    useRouter();
    const email = ref("");
    const password = ref("");
    const error = ref("");
    const loading = ref(false);
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "min-h-screen bg-[#0f172a] flex items-center justify-center p-4" }, _attrs))}><div class="w-full max-w-md"><div class="text-center mb-8"><div class="flex items-center justify-center gap-3 mb-4"><div class="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">`);
      _push(ssrRenderComponent(unref(Target), { class: "text-white w-7 h-7" }, null, _parent));
      _push(`</div><span class="text-3xl font-bold font-outfit text-white tracking-tight">VagAI</span></div><p class="text-slate-400">Entre na sua conta para continuar</p></div><div class="glass-card p-8"><form class="space-y-5"><div><label class="block text-sm font-medium text-slate-300 mb-2">Email</label><input${ssrRenderAttr("value", email.value)} type="email" required class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="seu@email.com"></div><div><label class="block text-sm font-medium text-slate-300 mb-2">Senha</label><input${ssrRenderAttr("value", password.value)} type="password" required class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="••••••••"></div>`);
      if (error.value) {
        _push(`<div class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">${ssrInterpolate(error.value)}</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<button type="submit"${ssrIncludeBooleanAttr(loading.value) ? " disabled" : ""} class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors flex items-center justify-center gap-2">`);
      if (loading.value) {
        _push(ssrRenderComponent(unref(Loader2), { class: "w-5 h-5 animate-spin" }, null, _parent));
      } else {
        _push(`<!---->`);
      }
      _push(` ${ssrInterpolate(loading.value ? "Entrando..." : "Entrar")}</button></form><div class="mt-6 text-center"><p class="text-slate-400 text-sm"> Não tem conta? `);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/register",
        class: "text-indigo-400 hover:text-indigo-300 font-medium"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(` Criar conta grátis `);
          } else {
            return [
              createTextVNode(" Criar conta grátis ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`</p></div></div></div></div>`);
    };
  }
};
const _sfc_setup$p = _sfc_main$p.setup;
_sfc_main$p.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/auth/Login.vue");
  return _sfc_setup$p ? _sfc_setup$p(props, ctx) : void 0;
};
const _sfc_main$o = {
  __name: "Register",
  __ssrInlineRender: true,
  setup(__props) {
    useRouter();
    const name = ref("");
    const email = ref("");
    const password = ref("");
    const error = ref("");
    const loading = ref(false);
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "min-h-screen bg-[#0f172a] flex items-center justify-center p-4" }, _attrs))}><div class="w-full max-w-md"><div class="text-center mb-8"><div class="flex items-center justify-center gap-3 mb-4"><div class="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">`);
      _push(ssrRenderComponent(unref(Target), { class: "text-white w-7 h-7" }, null, _parent));
      _push(`</div><span class="text-3xl font-bold font-outfit text-white tracking-tight">VagAI</span></div><p class="text-slate-400">Crie sua conta e comece a usar IA para encontrar vagas</p></div><div class="glass-card p-8"><form class="space-y-5"><div><label class="block text-sm font-medium text-slate-300 mb-2">Nome</label><input${ssrRenderAttr("value", name.value)} type="text" required minlength="2" class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="Seu nome"></div><div><label class="block text-sm font-medium text-slate-300 mb-2">Email</label><input${ssrRenderAttr("value", email.value)} type="email" required class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="seu@email.com"></div><div><label class="block text-sm font-medium text-slate-300 mb-2">Senha</label><input${ssrRenderAttr("value", password.value)} type="password" required minlength="8" class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="Mínimo 8 caracteres"></div>`);
      if (error.value) {
        _push(`<div class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">${ssrInterpolate(error.value)}</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<button type="submit"${ssrIncludeBooleanAttr(loading.value) ? " disabled" : ""} class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors flex items-center justify-center gap-2">`);
      if (loading.value) {
        _push(ssrRenderComponent(unref(Loader2), { class: "w-5 h-5 animate-spin" }, null, _parent));
      } else {
        _push(`<!---->`);
      }
      _push(` ${ssrInterpolate(loading.value ? "Criando conta..." : "Criar conta grátis")}</button></form><div class="mt-6 text-center"><p class="text-slate-400 text-sm"> Já tem conta? `);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/login",
        class: "text-indigo-400 hover:text-indigo-300 font-medium"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(` Fazer login `);
          } else {
            return [
              createTextVNode(" Fazer login ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`</p></div></div></div></div>`);
    };
  }
};
const _sfc_setup$o = _sfc_main$o.setup;
_sfc_main$o.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/auth/Register.vue");
  return _sfc_setup$o ? _sfc_setup$o(props, ctx) : void 0;
};
const _sfc_main$n = {
  __name: "ExperienceEntry",
  __ssrInlineRender: true,
  props: {
    modelValue: {
      type: Object,
      default: () => ({ company: "", role: "", start_date: "", end_date: "", description: "" })
    }
  },
  emits: ["update:modelValue", "remove"],
  setup(__props, { emit: __emit }) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "bg-slate-950/50 border border-white/5 rounded-2xl p-5 space-y-4 relative group" }, _attrs))}><div class="flex items-center justify-between"><h4 class="text-sm font-bold text-slate-400 uppercase tracking-widest">Experiencia</h4><button type="button" class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all opacity-0 group-hover:opacity-100">`);
      _push(ssrRenderComponent(unref(Trash2), { class: "w-4 h-4" }, null, _parent));
      _push(`</button></div><div class="grid grid-cols-1 md:grid-cols-2 gap-4"><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Empresa</label><input${ssrRenderAttr("value", __props.modelValue.company)} type="text" class="input-field w-full h-10 text-sm" placeholder="Nome da empresa"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Cargo</label><input${ssrRenderAttr("value", __props.modelValue.role)} type="text" class="input-field w-full h-10 text-sm" placeholder="Seu cargo"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Inicio</label><input${ssrRenderAttr("value", __props.modelValue.start_date)} type="text" class="input-field w-full h-10 text-sm" placeholder="MM/AAAA"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Fim</label><input${ssrRenderAttr("value", __props.modelValue.end_date)} type="text" class="input-field w-full h-10 text-sm" placeholder="MM/AAAA ou Presente"></div></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Descricao da atividade do cargo</label><textarea class="input-field w-full text-sm min-h-[80px] resize-y" placeholder="Descreva suas atividades e responsabilidades neste cargo..." rows="3">${ssrInterpolate(__props.modelValue.description)}</textarea></div></div>`);
    };
  }
};
const _sfc_setup$n = _sfc_main$n.setup;
_sfc_main$n.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/ExperienceEntry.vue");
  return _sfc_setup$n ? _sfc_setup$n(props, ctx) : void 0;
};
const _sfc_main$m = {
  __name: "EducationEntry",
  __ssrInlineRender: true,
  props: {
    modelValue: {
      type: Object,
      default: () => ({ institution: "", degree: "", field: "", start_date: "", end_date: "", notes: "" })
    }
  },
  emits: ["update:modelValue", "remove"],
  setup(__props, { emit: __emit }) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "bg-slate-950/50 border border-white/5 rounded-2xl p-5 space-y-4 relative group" }, _attrs))}><div class="flex items-center justify-between"><h4 class="text-sm font-bold text-slate-400 uppercase tracking-widest">Educacao</h4><button type="button" class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all opacity-0 group-hover:opacity-100">`);
      _push(ssrRenderComponent(unref(Trash2), { class: "w-4 h-4" }, null, _parent));
      _push(`</button></div><div class="grid grid-cols-1 md:grid-cols-2 gap-4"><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Instituicao</label><input${ssrRenderAttr("value", __props.modelValue.institution)} type="text" class="input-field w-full h-10 text-sm" placeholder="Nome da instituicao"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Grau</label><input${ssrRenderAttr("value", __props.modelValue.degree)} type="text" class="input-field w-full h-10 text-sm" placeholder="Graduacao, Mestrado, etc."></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Area de Estudo</label><input${ssrRenderAttr("value", __props.modelValue.field)} type="text" class="input-field w-full h-10 text-sm" placeholder="Ciencia da Computacao, Engenharia, etc."></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Inicio</label><input${ssrRenderAttr("value", __props.modelValue.start_date)} type="text" class="input-field w-full h-10 text-sm" placeholder="MM/AAAA"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Fim</label><input${ssrRenderAttr("value", __props.modelValue.end_date)} type="text" class="input-field w-full h-10 text-sm" placeholder="MM/AAAA"></div><div class="space-y-1 md:col-span-2"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Observacoes</label><input${ssrRenderAttr("value", __props.modelValue.notes)} type="text" class="input-field w-full h-10 text-sm" placeholder="Notas adicionais (opcional)"></div></div></div>`);
    };
  }
};
const _sfc_setup$m = _sfc_main$m.setup;
_sfc_main$m.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/EducationEntry.vue");
  return _sfc_setup$m ? _sfc_setup$m(props, ctx) : void 0;
};
const _sfc_main$l = {
  __name: "TagInput",
  __ssrInlineRender: true,
  props: {
    modelValue: { type: Array, default: () => [] },
    placeholder: { type: String, default: "Digite e pressione Enter" }
  },
  emits: ["update:modelValue"],
  setup(__props, { emit: __emit }) {
    const newTag = ref("");
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-2" }, _attrs))}><div class="flex flex-wrap gap-2"><!--[-->`);
      ssrRenderList(__props.modelValue, (tag, i) => {
        _push(`<span class="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-full text-sm text-indigo-300">${ssrInterpolate(tag)} <button type="button" class="w-4 h-4 flex items-center justify-center rounded-full bg-indigo-500/20 hover:bg-red-500/30 hover:text-red-400 transition-colors text-xs"> × </button></span>`);
      });
      _push(`<!--]--></div><div class="flex gap-2"><input${ssrRenderAttr("value", newTag.value)} type="text" class="input-field flex-1 h-10 text-sm"${ssrRenderAttr("placeholder", __props.placeholder)}><button type="button" class="px-4 h-10 bg-slate-700/50 hover:bg-indigo-500/20 text-slate-300 hover:text-indigo-400 rounded-xl text-sm font-bold transition-all"> Adicionar </button></div></div>`);
    };
  }
};
const _sfc_setup$l = _sfc_main$l.setup;
_sfc_main$l.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/TagInput.vue");
  return _sfc_setup$l ? _sfc_setup$l(props, ctx) : void 0;
};
const _sfc_main$k = {
  __name: "ResumeForm",
  __ssrInlineRender: true,
  props: {
    initialData: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ["save", "generate-pdf"],
  setup(__props, { expose: __expose, emit: __emit }) {
    const props = __props;
    const safeData = () => {
      const d = props.initialData || {};
      return {
        personal_info: d.personal_info || { name: "", email: "", phone: "", location: "", linkedin: "", website: "" },
        summary: d.summary || "",
        experience: Array.isArray(d.experience) ? d.experience : [],
        education: Array.isArray(d.education) ? d.education : [],
        skills: Array.isArray(d.skills) ? d.skills : [],
        languages: Array.isArray(d.languages) ? d.languages : [],
        certifications: Array.isArray(d.certifications) ? d.certifications : []
      };
    };
    const init = safeData();
    const form = reactive({
      personal_info: { ...init.personal_info },
      summary: init.summary,
      experience: [...init.experience],
      education: [...init.education],
      skills: [...init.skills],
      languages: [...init.languages],
      certifications: [...init.certifications]
    });
    watch(() => props.initialData, (newData) => {
      if (newData) {
        Object.assign(form.personal_info, newData.personal_info || {});
        form.summary = newData.summary || "";
        form.experience.splice(0, form.experience.length, ...newData.experience || []);
        form.education.splice(0, form.education.length, ...newData.education || []);
        form.skills.splice(0, form.skills.length, ...newData.skills || []);
        form.languages.splice(0, form.languages.length, ...newData.languages || []);
        form.certifications.splice(0, form.certifications.length, ...newData.certifications || []);
      }
    }, { deep: true });
    const removeExperience = (index) => {
      form.experience.splice(index, 1);
    };
    const removeEducation = (index) => {
      form.education.splice(index, 1);
    };
    const getFormData = () => ({ ...form });
    __expose({ getFormData });
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<form${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="glass-card p-8"><div class="flex items-center gap-3 mb-6"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
      _push(ssrRenderComponent(unref(User), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Informacoes Pessoais</h3></div><div class="grid grid-cols-1 md:grid-cols-2 gap-4"><div class="space-y-1 md:col-span-2"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Nome completo *</label><input${ssrRenderAttr("value", form.personal_info.name)} type="text" class="input-field w-full h-10 text-sm" placeholder="Seu nome" required></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Email</label><input${ssrRenderAttr("value", form.personal_info.email)} type="email" class="input-field w-full h-10 text-sm" placeholder="email@exemplo.com"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Telefone</label><input${ssrRenderAttr("value", form.personal_info.phone)} type="text" class="input-field w-full h-10 text-sm" placeholder="(11) 99999-9999"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Localizacao</label><input${ssrRenderAttr("value", form.personal_info.location)} type="text" class="input-field w-full h-10 text-sm" placeholder="Sao Paulo, SP"></div><div class="space-y-1"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">LinkedIn</label><input${ssrRenderAttr("value", form.personal_info.linkedin)} type="text" class="input-field w-full h-10 text-sm" placeholder="linkedin.com/in/seu-perfil"></div><div class="space-y-1 md:col-span-2"><label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Website</label><input${ssrRenderAttr("value", form.personal_info.website)} type="text" class="input-field w-full h-10 text-sm" placeholder="https://seusite.com"></div></div></div><div class="glass-card p-8"><div class="flex items-center gap-3 mb-6"><div class="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400">`);
      _push(ssrRenderComponent(unref(FileText), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Resumo / Objetivo</h3></div><textarea class="input-field w-full text-sm min-h-[100px] resize-y" placeholder="Descreva seu perfil profissional, objetivos e diferencias..." rows="4">${ssrInterpolate(form.summary)}</textarea></div><div class="glass-card p-8"><div class="flex items-center justify-between mb-6"><div class="flex items-center gap-3"><div class="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center text-blue-400">`);
      _push(ssrRenderComponent(unref(Briefcase), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Experiencia Profissional</h3></div><button type="button" class="px-4 py-2 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 rounded-xl text-sm font-bold transition-all flex items-center gap-2">`);
      _push(ssrRenderComponent(unref(Plus), { class: "w-4 h-4" }, null, _parent));
      _push(` Adicionar </button></div>`);
      if (form.experience.length === 0) {
        _push(`<div class="text-center py-8 text-slate-500 text-sm"> Nenhuma experiencia adicionada. Clique em &quot;Adicionar&quot; para comecar. </div>`);
      } else {
        _push(`<div class="space-y-4"><!--[-->`);
        ssrRenderList(form.experience, (exp, i) => {
          _push(ssrRenderComponent(_sfc_main$n, {
            key: i,
            "model-value": exp,
            "onUpdate:modelValue": ($event) => form.experience[i] = $event,
            onRemove: ($event) => removeExperience(i)
          }, null, _parent));
        });
        _push(`<!--]--></div>`);
      }
      _push(`</div><div class="glass-card p-8"><div class="flex items-center justify-between mb-6"><div class="flex items-center gap-3"><div class="w-10 h-10 bg-purple-500/10 rounded-xl flex items-center justify-center text-purple-400">`);
      _push(ssrRenderComponent(unref(GraduationCap), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Educacao</h3></div><button type="button" class="px-4 py-2 bg-purple-500/10 hover:bg-purple-500/20 text-purple-400 rounded-xl text-sm font-bold transition-all flex items-center gap-2">`);
      _push(ssrRenderComponent(unref(Plus), { class: "w-4 h-4" }, null, _parent));
      _push(` Adicionar </button></div>`);
      if (form.education.length === 0) {
        _push(`<div class="text-center py-8 text-slate-500 text-sm"> Nenhuma formacao adicionada. </div>`);
      } else {
        _push(`<div class="space-y-4"><!--[-->`);
        ssrRenderList(form.education, (edu, i) => {
          _push(ssrRenderComponent(_sfc_main$m, {
            key: i,
            "model-value": edu,
            "onUpdate:modelValue": ($event) => form.education[i] = $event,
            onRemove: ($event) => removeEducation(i)
          }, null, _parent));
        });
        _push(`<!--]--></div>`);
      }
      _push(`</div><div class="glass-card p-8"><div class="flex items-center gap-3 mb-6"><div class="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center text-amber-400">`);
      _push(ssrRenderComponent(unref(Wrench), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Habilidades</h3></div>`);
      _push(ssrRenderComponent(_sfc_main$l, {
        modelValue: form.skills,
        "onUpdate:modelValue": ($event) => form.skills = $event,
        placeholder: "Ex: JavaScript, Python, React..."
      }, null, _parent));
      _push(`</div><div class="glass-card p-8"><div class="flex items-center gap-3 mb-6"><div class="w-10 h-10 bg-cyan-500/10 rounded-xl flex items-center justify-center text-cyan-400">`);
      _push(ssrRenderComponent(unref(Globe), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Idiomas</h3></div>`);
      _push(ssrRenderComponent(_sfc_main$l, {
        modelValue: form.languages,
        "onUpdate:modelValue": ($event) => form.languages = $event,
        placeholder: "Ex: Portugues, Ingles, Espanhol..."
      }, null, _parent));
      _push(`</div><div class="glass-card p-8"><div class="flex items-center gap-3 mb-6"><div class="w-10 h-10 bg-rose-500/10 rounded-xl flex items-center justify-center text-rose-400">`);
      _push(ssrRenderComponent(unref(Award), { class: "w-5 h-5" }, null, _parent));
      _push(`</div><h3 class="text-lg font-bold text-white font-outfit">Certificacoes</h3></div>`);
      _push(ssrRenderComponent(_sfc_main$l, {
        modelValue: form.certifications,
        "onUpdate:modelValue": ($event) => form.certifications = $event,
        placeholder: "Ex: AWS Solutions Architect, PMP..."
      }, null, _parent));
      _push(`</div><div class="flex gap-4"><button type="submit" class="flex-1 h-12 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-emerald-900/20 active:scale-95 flex items-center justify-center gap-2">`);
      _push(ssrRenderComponent(unref(Save), { class: "w-5 h-5" }, null, _parent));
      _push(` Salvar Curriculo </button><button type="button" class="h-12 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2">`);
      _push(ssrRenderComponent(unref(Download), { class: "w-5 h-5" }, null, _parent));
      _push(` Gerar PDF </button></div></form>`);
    };
  }
};
const _sfc_setup$k = _sfc_main$k.setup;
_sfc_main$k.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/ResumeForm.vue");
  return _sfc_setup$k ? _sfc_setup$k(props, ctx) : void 0;
};
const STORAGE_KEY = "vagai_resume_draft";
const _sfc_main$j = {
  __name: "ResumeEditor",
  __ssrInlineRender: true,
  setup(__props) {
    const route = useRoute();
    const router2 = useRouter();
    const resumeFormRef = ref(null);
    const resumeId = ref(null);
    const resumeData = ref(null);
    const parseFile = ref(null);
    const parsing = ref(false);
    const saveMessage = ref("");
    const errorMsg = ref("");
    const loadFromStorage = () => {
      try {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) {
          const parsed = JSON.parse(saved);
          if (parsed && parsed.personal_info) {
            return parsed;
          }
        }
      } catch (e) {
      }
      return null;
    };
    const saveToStorage = (data) => {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
      } catch (e) {
      }
    };
    onMounted(async () => {
      const id = route.params.id;
      if (id && id !== "new") {
        try {
          const result = await getResumeData(id);
          if (result && result.data) {
            resumeId.value = result.resume_id;
            resumeData.value = result.data;
          }
        } catch (e) {
          console.error("Erro ao carregar curriculo:", e);
          errorMsg.value = "Erro ao carregar curriculo do servidor.";
        }
      } else {
        const draft = loadFromStorage();
        if (draft) {
          resumeData.value = draft;
        }
      }
    });
    let autoSaveInterval = null;
    onMounted(() => {
      autoSaveInterval = setInterval(() => {
        if (resumeFormRef.value && typeof resumeFormRef.value.getFormData === "function") {
          saveToStorage(resumeFormRef.value.getFormData());
        }
      }, 3e4);
    });
    onUnmounted(() => {
      if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
      }
    });
    const navigateToCreated = (id) => {
      if (route.params.id === "new" || !route.params.id) {
        router2.replace({ params: { id: String(id) } });
      }
    };
    const ensureResumeSaved = async (data) => {
      var _a;
      if (resumeId.value) {
        await updateResumeData(resumeId.value, data);
        return resumeId.value;
      }
      const response = await createResume(data);
      const body = response.data || response;
      resumeId.value = ((_a = body.resume) == null ? void 0 : _a.id) || null;
      if (!resumeId.value) {
        throw new Error("Falha ao criar curriculo no servidor");
      }
      navigateToCreated(resumeId.value);
      return resumeId.value;
    };
    const handleSave = async (data) => {
      var _a, _b;
      try {
        await ensureResumeSaved(data);
        saveToStorage(data);
        showSaveMessage("Curriculo salvo com sucesso!");
      } catch (e) {
        console.error("Erro ao salvar:", e);
        saveToStorage(data);
        showSaveMessage(((_b = (_a = e.response) == null ? void 0 : _a.data) == null ? void 0 : _b.error) || "Salvo localmente (erro ao sincronizar)");
      }
    };
    const handleGeneratePDF = async () => {
      var _a, _b;
      try {
        let id = resumeId.value;
        if (!id && resumeFormRef.value && typeof resumeFormRef.value.getFormData === "function") {
          id = await ensureResumeSaved(resumeFormRef.value.getFormData());
        }
        if (!id) {
          errorMsg.value = "Salve o curriculo antes de gerar o PDF.";
          return;
        }
        const response = await generateResumePDF(id);
        const blob = new Blob([response.data], { type: "application/pdf" });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "curriculo.pdf";
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      } catch (e) {
        console.error("Erro ao gerar PDF:", e);
        errorMsg.value = ((_b = (_a = e.response) == null ? void 0 : _a.data) == null ? void 0 : _b.error) || "Erro ao gerar PDF. Tente novamente.";
      }
    };
    const showSaveMessage = (msg) => {
      saveMessage.value = msg;
      setTimeout(() => {
        saveMessage.value = "";
      }, 3e3);
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="flex items-end justify-between gap-4"><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Criar seu Curriculo</h1><p class="text-slate-400">Importe um curriculo existente ou comece do zero. Edite todos os campos e gere um PDF profissional.</p>`);
      if (resumeId.value) {
        _push(`<span class="inline-flex items-center gap-1.5 mt-3 px-2.5 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-xs text-emerald-300">`);
        _push(ssrRenderComponent(unref(CircleCheck), { class: "w-3.5 h-3.5" }, null, _parent));
        _push(` Salvo no servidor </span>`);
      } else {
        _push(`<span class="inline-flex items-center gap-1.5 mt-3 px-2.5 py-1 bg-amber-500/10 border border-amber-500/20 rounded-lg text-xs text-amber-300">`);
        _push(ssrRenderComponent(unref(CircleAlert), { class: "w-3.5 h-3.5" }, null, _parent));
        _push(` Não salvo no servidor </span>`);
      }
      _push(`</div></div>`);
      if (errorMsg.value) {
        _push(`<div class="glass-card p-6 border-red-500/30 bg-red-500/5"><p class="text-red-400 text-sm">${ssrInterpolate(errorMsg.value)}</p><button class="text-xs text-slate-500 mt-2 hover:text-white">Fechar</button></div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<div class="glass-card p-10"><div class="flex items-center gap-3 mb-8"><div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">`);
      _push(ssrRenderComponent(unref(Upload), { class: "w-6 h-6" }, null, _parent));
      _push(`</div><div><h2 class="text-2xl font-bold text-white font-outfit">Importar Curriculo</h2><p class="text-sm text-slate-400">${ssrInterpolate(resumeData.value ? "Atualize o curriculo atual importando um novo arquivo" : "Importe um curriculo existente para comecar")}</p></div></div><form class="space-y-6"><div class="group relative h-48 border-2 border-dashed border-white/10 hover:border-indigo-500/50 rounded-2xl flex flex-col items-center justify-center transition-all duration-300 cursor-pointer overflow-hidden"><input type="file" class="absolute inset-0 opacity-0 cursor-pointer z-10" accept=".pdf,.txt,.docx"><div class="flex flex-col items-center gap-3 group-hover:scale-110 transition-transform duration-500"><div class="w-12 h-12 bg-slate-900 rounded-full flex items-center justify-center text-slate-500">`);
      _push(ssrRenderComponent(unref(Upload), { class: "w-6 h-6" }, null, _parent));
      _push(`</div><div class="text-center"><p class="text-white font-bold">${ssrInterpolate(parseFile.value ? parseFile.value.name : "Clique ou arraste seu arquivo")}</p><p class="text-xs text-slate-500 mt-1">PDF, TXT ou DOCX (Max. 10MB)</p></div></div></div><button type="submit" class="w-full h-12 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2"${ssrIncludeBooleanAttr(!parseFile.value || parsing.value) ? " disabled" : ""}>`);
      if (parsing.value) {
        _push(`<div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>`);
      } else {
        _push(ssrRenderComponent(unref(FileSearch), { class: "w-5 h-5" }, null, _parent));
      }
      _push(` ${ssrInterpolate(parsing.value ? "Analisando curriculo com IA..." : "Parsear e Editar")}</button></form>`);
      if (!resumeData.value) {
        _push(`<div class="mt-6 text-center"><button class="text-sm text-slate-500 hover:text-indigo-400 transition-colors underline underline-offset-4"> Ou comece do zero </button></div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
      if (resumeData.value) {
        _push(`<div class="space-y-6">`);
        _push(ssrRenderComponent(_sfc_main$k, {
          ref_key: "resumeFormRef",
          ref: resumeFormRef,
          "initial-data": resumeData.value,
          onSave: handleSave,
          onGeneratePdf: handleGeneratePDF
        }, null, _parent));
        if (saveMessage.value) {
          _push(`<div class="fixed bottom-6 right-6 bg-emerald-600 text-white px-6 py-3 rounded-xl shadow-lg z-50">${ssrInterpolate(saveMessage.value)}</div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$j = _sfc_main$j.setup;
_sfc_main$j.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/ResumeEditor.vue");
  return _sfc_setup$j ? _sfc_setup$j(props, ctx) : void 0;
};
const REVIEW_DAYS = 3;
const _sfc_main$i = {
  __name: "InterviewPrep",
  __ssrInlineRender: true,
  setup(__props) {
    const route = useRoute();
    const prep = ref(null);
    const questions = ref([]);
    const isLoading = ref(true);
    const regenerating = ref(false);
    const verifying = ref(false);
    const currentIndex = ref(0);
    const revealed = ref(false);
    const draftAnswer = ref("");
    const simMinutes = ref(0);
    const timeLeft = ref(0);
    const questionStartAt = ref(Date.now());
    let simTimer = null;
    const fmtClock = (s) => `${Math.floor(Math.max(0, s) / 60)}:${String(Math.max(0, s) % 60).padStart(2, "0")}`;
    const questionElapsed = () => Math.max(0, Math.floor((Date.now() - questionStartAt.value) / 1e3));
    const stopSimTimer = () => {
      if (simTimer) {
        clearInterval(simTimer);
        simTimer = null;
      }
    };
    const startSimTimer = () => {
      stopSimTimer();
      questionStartAt.value = Date.now();
      if (simMinutes.value <= 0 || questions.value.length === 0) return;
      timeLeft.value = simMinutes.value * 60;
      simTimer = setInterval(() => {
        timeLeft.value--;
        if (timeLeft.value <= 0) {
          stopSimTimer();
          if (currentIndex.value < questions.value.length - 1) next();
        }
      }, 1e3);
    };
    const current = computed(() => questions.value[currentIndex.value] || {});
    const verification = computed(() => {
      var _a;
      return ((_a = prep.value) == null ? void 0 : _a.verification) || null;
    });
    const historyAsc = computed(() => {
      var _a;
      return [...((_a = prep.value) == null ? void 0 : _a.history) || []].reverse();
    });
    const currentAnalysis = computed(() => {
      var _a, _b, _c;
      const a = (_b = (_a = prep.value) == null ? void 0 : _a.analyses) == null ? void 0 : _b[current.value.id];
      if (!a || a.status !== "completed" || !((_c = a.text) == null ? void 0 : _c.trim())) return null;
      return a;
    });
    const progress = computed(() => {
      var _a;
      return ((_a = prep.value) == null ? void 0 : _a.progress) || {};
    });
    const remainingCount = computed(() => progress.value.remaining ?? 0);
    const answeredCount = computed(() => progress.value.answered ?? 0);
    const verificationStatusLabel = computed(() => {
      var _a, _b;
      return {
        verified: "Verificada",
        running: "Verificando...",
        pending: "Na fila",
        outdated: "Desatualizada"
      }[(_a = verification.value) == null ? void 0 : _a.status] || ((_b = verification.value) == null ? void 0 : _b.status) || "";
    });
    const verificationStatusClass = computed(() => {
      var _a;
      return {
        verified: "bg-emerald-500/15 text-emerald-300",
        running: "bg-indigo-500/15 text-indigo-300",
        pending: "bg-slate-500/15 text-slate-300",
        outdated: "bg-amber-500/15 text-amber-300"
      }[(_a = verification.value) == null ? void 0 : _a.status] || "bg-white/5 text-slate-400";
    });
    const masteredCount = computed(() => questions.value.filter((q) => q.status === "mastered").length);
    const practicedCount = computed(() => questions.value.filter((q) => q.status === "practiced").length);
    const entrySpent = computed(() => {
      var _a, _b, _c;
      return ((_c = (_b = (_a = prep.value) == null ? void 0 : _a.entries) == null ? void 0 : _b[current.value.id]) == null ? void 0 : _c.time_spent_seconds) || 0;
    });
    const reviewQueue = computed(() => {
      const out = [];
      const now = Date.now();
      questions.value.forEach((q, i) => {
        var _a, _b;
        const e = (_b = (_a = prep.value) == null ? void 0 : _a.entries) == null ? void 0 : _b[q.id];
        if ((e == null ? void 0 : e.self_assessment) === "low" && (e == null ? void 0 : e.updated_at)) {
          const days = Math.floor((now - new Date(e.updated_at).getTime()) / 864e5);
          if (days >= REVIEW_DAYS) out.push({ index: i, q, days });
        }
      });
      return out;
    });
    const progressPercent = computed(() => {
      if (questions.value.length === 0) return 0;
      return Math.round((masteredCount.value + practicedCount.value) / questions.value.length * 100);
    });
    const applyPrep = (data) => {
      var _a, _b, _c;
      prep.value = data.preparation;
      questions.value = ((_a = data.preparation) == null ? void 0 : _a.questions) || [];
      const firstPending = questions.value.findIndex((q) => q.status !== "mastered");
      currentIndex.value = firstPending === -1 ? 0 : firstPending;
      revealed.value = false;
      const currentQ = questions.value[currentIndex.value];
      if (currentQ && ((_b = prep.value) == null ? void 0 : _b.entries)) {
        draftAnswer.value = ((_c = prep.value.entries[currentQ.id]) == null ? void 0 : _c.answer) || "";
      } else {
        draftAnswer.value = "";
      }
      startPolling();
      startSimTimer();
    };
    let pollTimer = null;
    const stopPolling = () => {
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
    };
    const startPolling = () => {
      var _a;
      const st = (_a = verification.value) == null ? void 0 : _a.status;
      if (st === "pending" || st === "running") {
        stopPolling();
        pollTimer = setInterval(async () => {
          var _a2;
          await refreshMeta();
          const st2 = (_a2 = verification.value) == null ? void 0 : _a2.status;
          if (!st2 || st2 === "verified" || st2 === "outdated") stopPolling();
        }, 2500);
      } else {
        stopPolling();
      }
    };
    const refreshMeta = async () => {
      var _a;
      try {
        const data = await getInterviewPrep(route.params.id);
        prep.value = data.preparation;
        questions.value = ((_a = data.preparation) == null ? void 0 : _a.questions) || [];
      } catch (e) {
        console.error("Falha ao atualizar preparação", e);
      }
    };
    onUnmounted(() => {
      stopPolling();
      stopSimTimer();
    });
    const load = async () => {
      isLoading.value = true;
      try {
        const data = await getInterviewPrep(route.params.id);
        applyPrep(data);
      } catch (e) {
        console.error("Falha ao carregar preparação", e);
      } finally {
        isLoading.value = false;
      }
    };
    load();
    watch(() => route.params.id, load);
    watch(currentIndex, () => {
      var _a, _b;
      revealed.value = false;
      const currentQ = questions.value[currentIndex.value];
      if (currentQ && ((_a = prep.value) == null ? void 0 : _a.entries)) {
        draftAnswer.value = ((_b = prep.value.entries[currentQ.id]) == null ? void 0 : _b.answer) || "";
      } else {
        draftAnswer.value = "";
      }
      startSimTimer();
    });
    watch(simMinutes, () => {
      startSimTimer();
    });
    const next = async () => {
      if (currentIndex.value < questions.value.length - 1) {
        await saveDraft();
        currentIndex.value++;
      }
    };
    const saveDraft = async () => {
      var _a, _b, _c, _d;
      const q = current.value;
      if (!q || !prep.value) return;
      try {
        const currentSelfAssessment = ((_b = (_a = prep.value.entries) == null ? void 0 : _a[q.id]) == null ? void 0 : _b.self_assessment) || "";
        if (!prep.value.entries) prep.value.entries = {};
        const wasEmpty = !(((_c = prep.value.entries[q.id]) == null ? void 0 : _c.answer) || "").trim();
        const isNowEmpty = !draftAnswer.value.trim();
        prep.value.entries[q.id] = {
          id: (_d = prep.value.entries[q.id]) == null ? void 0 : _d.id,
          question_id: q.id,
          answer: draftAnswer.value,
          self_assessment: currentSelfAssessment
        };
        if (prep.value.progress) {
          if (wasEmpty && !isNowEmpty) prep.value.progress.answered++;
          else if (!wasEmpty && isNowEmpty) prep.value.progress.answered--;
        }
        await saveQuestionAnswer(prep.value.id, q.id, draftAnswer.value, currentSelfAssessment, questionElapsed());
      } catch (e) {
        console.error("Falha ao salvar rascunho", e);
      }
    };
    const categoryLabel = (c) => ({ technology: "Tecnologia", foundations: "Fundamentos", architecture: "Arquitetura" })[c] || c;
    const categoryClass = (c) => ({
      technology: "bg-indigo-500/15 text-indigo-300 border border-indigo-500/30",
      foundations: "bg-amber-500/15 text-amber-300 border border-amber-500/30",
      architecture: "bg-emerald-500/15 text-emerald-300 border border-emerald-500/30"
    })[c] || "bg-white/5 text-slate-400";
    const dotClass = (q, i) => {
      if (i === currentIndex.value) return "bg-indigo-600 text-white ring-2 ring-indigo-400/40";
      if (q.status === "mastered") return "bg-emerald-500/20 text-emerald-300";
      if (q.status === "practiced") return "bg-amber-500/20 text-amber-300";
      return "bg-slate-800 text-slate-400 hover:bg-slate-700";
    };
    return (_ctx, _push, _parent, _attrs) => {
      var _a, _b, _c, _d, _e, _f, _g;
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div class="flex items-center justify-between"><div>`);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/applied",
        class: "inline-flex items-center gap-1.5 text-slate-400 hover:text-slate-200 text-sm mb-2 transition-colors"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(ArrowLeft), { class: "w-4 h-4" }, null, _parent2, _scopeId));
            _push2(` Candidatadas `);
          } else {
            return [
              createVNode(unref(ArrowLeft), { class: "w-4 h-4" }),
              createTextVNode(" Candidatadas ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`<h1 class="text-4xl font-bold text-white font-outfit tracking-tight">Sabatina</h1><p class="text-slate-400 mt-1"><span class="text-slate-300 font-medium">${ssrInterpolate((_a = prep.value) == null ? void 0 : _a.title)}</span>`);
      if ((_b = prep.value) == null ? void 0 : _b.company) {
        _push(`<span class="text-slate-500"> · ${ssrInterpolate(prep.value.company)}</span>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</p></div>`);
      if ((_c = prep.value) == null ? void 0 : _c.match_id) {
        _push(`<button${ssrIncludeBooleanAttr(regenerating.value) ? " disabled" : ""} class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-slate-800 border border-white/10 text-slate-300 text-sm font-semibold hover:bg-slate-700 transition-all disabled:opacity-40">`);
        if (regenerating.value) {
          _push(`<span class="w-4 h-4 border-2 border-slate-400/30 border-t-slate-300 rounded-full animate-spin"></span>`);
        } else {
          _push(ssrRenderComponent(unref(RefreshCw), { class: "w-4 h-4" }, null, _parent));
        }
        _push(` Regenerar </button>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
      if (isLoading.value) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (!prep.value || questions.value.length === 0) {
        _push(`<div class="flex flex-col items-center justify-center h-64 text-center"><div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">`);
        _push(ssrRenderComponent(unref(GraduationCap), { class: "w-10 h-10 text-slate-600" }, null, _parent));
        _push(`</div><p class="text-slate-500 max-w-md">Nenhuma pergunta disponível.</p></div>`);
      } else {
        _push(`<!--[--><div class="glass-card p-6"><div class="flex items-center justify-between mb-3"><span class="text-xs uppercase tracking-widest font-bold text-slate-400">Progresso</span><span class="text-sm font-mono text-slate-300">${ssrInterpolate(answeredCount.value)}/${ssrInterpolate(questions.value.length)} respondidas · ${ssrInterpolate(masteredCount.value)} dominadas · ${ssrInterpolate(practicedCount.value)} praticadas</span></div><div class="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden"><div class="h-full bg-gradient-to-r from-indigo-500 to-emerald-500 transition-all duration-500" style="${ssrRenderStyle({ width: progressPercent.value + "%" })}"></div></div></div>`);
        if (reviewQueue.value.length > 0) {
          _push(`<div class="glass-card p-6 border-amber-500/30"><div class="flex items-center justify-between mb-3"><span class="text-xs uppercase tracking-widest font-bold text-amber-300">Fila de revisão (${ssrInterpolate(reviewQueue.value.length)})</span><span class="text-[10px] text-slate-500">Marcadas como difíceis há 3+ dias</span></div><div class="flex flex-wrap gap-2"><!--[-->`);
          ssrRenderList(reviewQueue.value, (item) => {
            _push(`<button class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-200 text-xs font-semibold hover:bg-amber-500/20 transition-all"><span class="font-mono">#${ssrInterpolate(item.index + 1)}</span><span class="max-w-56 truncate">${ssrInterpolate(item.q.topic || item.q.text)}</span><span class="font-mono text-amber-400/80">${ssrInterpolate(item.days)}d</span></button>`);
          });
          _push(`<!--]--></div></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<div class="glass-card p-6 border-indigo-500/30"><div class="flex items-center justify-between mb-3"><span class="text-xs uppercase tracking-widest font-bold text-slate-400">Verificação por IA</span>`);
        if (verification.value) {
          _push(`<span class="${ssrRenderClass([verificationStatusClass.value, "px-2.5 py-1 rounded-lg text-[10px] uppercase tracking-widest font-bold"])}">${ssrInterpolate(verificationStatusLabel.value)}</span>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
        if (((_d = verification.value) == null ? void 0 : _d.status) === "verified") {
          _push(`<div class="flex flex-col sm:flex-row gap-6 items-start sm:items-center"><div class="text-center shrink-0"><div class="text-5xl font-bold font-outfit text-indigo-400">${ssrInterpolate(verification.value.score.toFixed(1))}</div><div class="text-[10px] uppercase tracking-widest font-bold text-slate-500 mt-1">Nota geral</div></div><div class="flex-1"><p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">${ssrInterpolate(verification.value.feedback)}</p><p class="text-[10px] text-slate-500 mt-2 font-mono">${ssrInterpolate(verification.value.source === "ai" ? "Avaliação por IA" : "Avaliação automática")} `);
          if (verification.value.verified_at) {
            _push(`<span> · ${ssrInterpolate(new Date(verification.value.verified_at).toLocaleString("pt-BR"))}</span>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</p>`);
          if (historyAsc.value.length > 1) {
            _push(`<div class="flex items-center gap-2 mt-3 flex-wrap"><span class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Evolução:</span><!--[-->`);
            ssrRenderList(historyAsc.value, (h, i) => {
              _push(`<span class="inline-flex items-center gap-1.5"><span class="px-2 py-0.5 rounded-lg bg-white/5 border border-white/10 text-xs font-mono text-slate-300"${ssrRenderAttr("title", new Date(h.verified_at).toLocaleString("pt-BR"))}>${ssrInterpolate(Number(h.score).toFixed(1))}</span>`);
              if (i < historyAsc.value.length - 1) {
                _push(`<span class="text-slate-600 text-xs">→</span>`);
              } else {
                _push(`<!---->`);
              }
              _push(`</span>`);
            });
            _push(`<!--]--></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div></div>`);
        } else if (((_e = verification.value) == null ? void 0 : _e.status) === "running" || ((_f = verification.value) == null ? void 0 : _f.status) === "pending") {
          _push(`<div class="flex items-center gap-4 py-2"><div class="w-8 h-8 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin shrink-0"></div><div><p class="text-slate-300 text-sm font-semibold">Verificando suas respostas...</p><p class="text-slate-500 text-xs mt-0.5">A nota e o feedback aparecerão aqui quando prontos. Você pode fechar esta página.</p></div></div>`);
        } else if (((_g = verification.value) == null ? void 0 : _g.status) === "outdated") {
          _push(`<div class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between"><p class="text-amber-300 text-sm font-semibold">Suas respostas mudaram após a verificação. Gere uma nova para atualizar a nota e o feedback.</p><button${ssrIncludeBooleanAttr(verifying.value) ? " disabled" : ""} class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold disabled:opacity-40 transition-all shrink-0">`);
          if (verifying.value) {
            _push(`<span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>`);
          } else {
            _push(ssrRenderComponent(unref(RefreshCw), { class: "w-4 h-4" }, null, _parent));
          }
          _push(` Gerar nova verificação </button></div>`);
        } else if (answeredCount.value === 0) {
          _push(`<div class="text-slate-500 text-sm"> Responda ao menos uma questão para pedir a correção por IA (${ssrInterpolate(progress.value.answered || 0)}/${ssrInterpolate(progress.value.total || 0)} respondidas). </div>`);
        } else {
          _push(`<div class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between"><p class="text-slate-400 text-sm">`);
          if (remainingCount.value > 0) {
            _push(`<span>Correção parcial: o que já foi respondido será corrigido — faltam <span class="text-slate-300 font-semibold">${ssrInterpolate(remainingCount.value)}</span>.</span>`);
          } else {
            _push(`<span>Todas as questões respondidas. Gere a verificação para receber sua nota e feedback.</span>`);
          }
          _push(`</p><button${ssrIncludeBooleanAttr(verifying.value) ? " disabled" : ""} class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold disabled:opacity-40 transition-all shrink-0">`);
          if (verifying.value) {
            _push(`<span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>`);
          } else {
            _push(ssrRenderComponent(unref(Sparkles), { class: "w-4 h-4" }, null, _parent));
          }
          _push(` Verificar minhas respostas </button></div>`);
        }
        _push(`</div><div class="flex items-center justify-center gap-1.5 flex-wrap"><!--[-->`);
        ssrRenderList(questions.value, (q, i) => {
          _push(`<button class="${ssrRenderClass([dotClass(q, i), "w-8 h-8 rounded-lg text-xs font-bold transition-all"])}"${ssrRenderAttr("title", "Pergunta " + (i + 1))}>${ssrInterpolate(i + 1)}</button>`);
        });
        _push(`<!--]--></div><div class="glass-card p-8 min-h-[24rem] flex flex-col"><div class="flex items-center gap-3 mb-6"><span class="${ssrRenderClass([categoryClass(current.value.category), "px-2.5 py-1 rounded-lg text-[10px] uppercase tracking-widest font-bold"])}">${ssrInterpolate(categoryLabel(current.value.category))}</span>`);
        if (current.value.topic) {
          _push(`<span class="px-2.5 py-1 rounded-lg bg-white/5 text-slate-400 text-[10px] uppercase tracking-widest font-bold">${ssrInterpolate(current.value.topic)}</span>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<span class="ml-auto text-xs text-slate-500 font-mono">${ssrInterpolate(currentIndex.value + 1)} / ${ssrInterpolate(questions.value.length)}</span><select title="Modo simulado: tempo por pergunta" class="bg-slate-900/60 border border-white/10 rounded-lg px-2 py-1 text-[10px] uppercase tracking-widest font-bold text-slate-400 focus:outline-none focus:border-indigo-500/50"><option${ssrRenderAttr("value", 0)}${ssrIncludeBooleanAttr(Array.isArray(simMinutes.value) ? ssrLooseContain(simMinutes.value, 0) : ssrLooseEqual(simMinutes.value, 0)) ? " selected" : ""}>Simulado off</option><option${ssrRenderAttr("value", 2)}${ssrIncludeBooleanAttr(Array.isArray(simMinutes.value) ? ssrLooseContain(simMinutes.value, 2) : ssrLooseEqual(simMinutes.value, 2)) ? " selected" : ""}>Simulado 2min</option><option${ssrRenderAttr("value", 3)}${ssrIncludeBooleanAttr(Array.isArray(simMinutes.value) ? ssrLooseContain(simMinutes.value, 3) : ssrLooseEqual(simMinutes.value, 3)) ? " selected" : ""}>Simulado 3min</option><option${ssrRenderAttr("value", 5)}${ssrIncludeBooleanAttr(Array.isArray(simMinutes.value) ? ssrLooseContain(simMinutes.value, 5) : ssrLooseEqual(simMinutes.value, 5)) ? " selected" : ""}>Simulado 5min</option></select>`);
        if (simMinutes.value > 0) {
          _push(`<span class="${ssrRenderClass([timeLeft.value <= 30 ? "bg-rose-500/15 text-rose-300" : "bg-white/5 text-slate-300", "px-2.5 py-1 rounded-lg text-[10px] font-mono font-bold"])}">${ssrInterpolate(fmtClock(timeLeft.value))}</span>`);
        } else {
          _push(`<!---->`);
        }
        if (entrySpent.value > 0) {
          _push(`<span class="text-[10px] text-slate-500 font-mono"${ssrRenderAttr("title", "Tempo total nesta pergunta")}>${ssrInterpolate(fmtClock(entrySpent.value))}</span>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div><h2 class="text-2xl font-bold text-white font-outfit leading-relaxed mb-6">${ssrInterpolate(current.value.text)}</h2>`);
        if (revealed.value) {
          _push(`<div class="bg-slate-900/70 border border-white/5 rounded-2xl p-6 mb-6"><div class="flex items-center gap-2 mb-3 text-indigo-400 font-bold text-xs uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(Lightbulb), { class: "w-4 h-4" }, null, _parent));
          _push(` Guia de resposta </div><p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">${ssrInterpolate(current.value.answer_guide || "Sem guia disponível.")}</p></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<div class="mb-6"><textarea rows="4" placeholder="Opcional: escreva sua resposta..." class="w-full bg-slate-900/60 border border-white/10 rounded-2xl p-4 text-slate-200 text-sm resize-none focus:outline-none focus:border-indigo-500/50 transition-colors">${ssrInterpolate(draftAnswer.value)}</textarea></div>`);
        if (currentAnalysis.value) {
          _push(`<div class="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6 mb-6"><div class="flex items-center gap-2 mb-3 text-emerald-300 font-bold text-xs uppercase tracking-widest">`);
          _push(ssrRenderComponent(unref(Sparkles), { class: "w-4 h-4" }, null, _parent));
          _push(` Feedback da correção por IA </div><p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">${ssrInterpolate(currentAnalysis.value.text)}</p></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<div class="mt-auto flex flex-col sm:flex-row items-center justify-between gap-4"><button${ssrIncludeBooleanAttr(currentIndex.value === 0) ? " disabled" : ""} class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all">`);
        _push(ssrRenderComponent(unref(ChevronLeft), { class: "w-4 h-4" }, null, _parent));
        _push(` Anterior </button>`);
        if (!revealed.value) {
          _push(`<div class="flex-1 flex justify-center"><button class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95">`);
          _push(ssrRenderComponent(unref(Eye), { class: "w-4 h-4" }, null, _parent));
          _push(` Revelar resposta </button></div>`);
        } else {
          _push(`<div class="flex-1 flex items-center justify-center gap-2"><button class="px-4 py-2.5 rounded-xl bg-rose-500/15 border border-rose-500/30 text-rose-300 text-sm font-semibold hover:bg-rose-500/25 transition-all">Difícil</button><button class="px-4 py-2.5 rounded-xl bg-amber-500/15 border border-amber-500/30 text-amber-300 text-sm font-semibold hover:bg-amber-500/25 transition-all">Ok</button><button class="px-4 py-2.5 rounded-xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 text-sm font-semibold hover:bg-emerald-500/25 transition-all">Dominada</button></div>`);
        }
        _push(`<button${ssrIncludeBooleanAttr(currentIndex.value === questions.value.length - 1) ? " disabled" : ""} class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all"> Próxima `);
        _push(ssrRenderComponent(unref(ChevronRight), { class: "w-4 h-4" }, null, _parent));
        _push(`</button></div></div><!--]-->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$i = _sfc_main$i.setup;
_sfc_main$i.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/InterviewPrep.vue");
  return _sfc_setup$i ? _sfc_setup$i(props, ctx) : void 0;
};
const _sfc_main$h = {
  __name: "InterviewPreparations",
  __ssrInlineRender: true,
  setup(__props) {
    const queryClient = useQueryClient();
    const { data: prepsResponse, isLoading } = useInterviewPreps();
    const preps = computed(() => {
      var _a;
      return ((_a = prepsResponse.value) == null ? void 0 : _a.data) || [];
    });
    const verifyingId = ref(null);
    let pollTimer = null;
    const stopPolling = () => {
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
    };
    const hasActiveVerification = computed(() => preps.value.some((p) => {
      var _a;
      return ["running", "pending"].includes((_a = p.verification) == null ? void 0 : _a.status);
    }));
    watch(hasActiveVerification, (active) => {
      if (active) {
        stopPolling();
        pollTimer = setInterval(() => queryClient.invalidateQueries({ queryKey: ["interview-preps"] }), 3e3);
      } else {
        stopPolling();
      }
    }, { immediate: true });
    onUnmounted(stopPolling);
    const answeredCount = (prep) => {
      const p = prep.progress;
      return p ? p.answered || 0 : 0;
    };
    const progressPercent = (prep) => {
      var _a;
      const total = ((_a = prep.progress) == null ? void 0 : _a.total) || 0;
      if (total === 0) return 0;
      return Math.round(answeredCount(prep) / total * 100);
    };
    const isCompleted = (prep) => {
      const p = prep.progress;
      return !!(p && p.total > 0 && (p.answered || 0) >= p.total);
    };
    const verification = (prep) => prep.verification || null;
    const canVerify = (prep) => isCompleted(prep);
    const isVerifying = (prep) => {
      var _a;
      if (verifyingId.value === prep.id) return true;
      const st = (_a = verification(prep)) == null ? void 0 : _a.status;
      return st === "running" || st === "pending";
    };
    const verifyLabel = () => "Verificar Respostas";
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Preparação para Entrevista</h1><p class="text-slate-400">Suas preparações de entrevista por vaga candidatada, com acompanhamento de progresso.</p></div><div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"><div></div>`);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/interview-prep/avulsa",
        class: "inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Wand2), { class: "w-4 h-4" }, null, _parent2, _scopeId));
            _push2(` Preparação Avulsa `);
          } else {
            return [
              createVNode(unref(Wand2), { class: "w-4 h-4" }),
              createTextVNode(" Preparação Avulsa ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`</div>`);
      if (unref(isLoading)) {
        _push(`<div class="flex items-center justify-center h-64"><div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (preps.value.length === 0) {
        _push(`<div class="flex flex-col items-center justify-center h-64 text-center"><div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">`);
        _push(ssrRenderComponent(unref(GraduationCap), { class: "w-10 h-10 text-slate-600" }, null, _parent));
        _push(`</div><p class="text-slate-500 max-w-md mb-6">Nenhuma preparação criada ainda. Gere uma a partir de uma vaga candidatada.</p>`);
        _push(ssrRenderComponent(_component_router_link, {
          to: "/applied",
          class: "inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95"
        }, {
          default: withCtx((_, _push2, _parent2, _scopeId) => {
            if (_push2) {
              _push2(ssrRenderComponent(unref(GraduationCap), { class: "w-4 h-4" }, null, _parent2, _scopeId));
              _push2(` Ver candidatadas `);
            } else {
              return [
                createVNode(unref(GraduationCap), { class: "w-4 h-4" }),
                createTextVNode(" Ver candidatadas ")
              ];
            }
          }),
          _: 1
        }, _parent));
        _push(`</div>`);
      } else {
        _push(`<div class="grid grid-cols-1 gap-6"><!--[-->`);
        ssrRenderList(preps.value, (prep) => {
          var _a, _b, _c, _d, _e, _f;
          _push(`<div class="glass-card p-8 border-l-4 border-indigo-500"><div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6"><div class="flex-1"><div class="flex items-center gap-3 mb-2"><h3 class="text-xl font-bold text-white font-outfit">${ssrInterpolate(prep.title)}</h3><span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">${ssrInterpolate(prep.id)}</span></div><div class="flex items-center gap-4 text-slate-400 text-sm"><span class="flex items-center gap-1.5">`);
          _push(ssrRenderComponent(unref(Building2), { class: "w-4 h-4" }, null, _parent));
          _push(` ${ssrInterpolate(prep.company || "Sem empresa")}</span><span class="w-1 h-1 bg-slate-700 rounded-full"></span><span class="px-2 py-0.5 rounded bg-white/5 text-slate-500 text-[10px] font-mono">${ssrInterpolate(prep.source === "ai" ? "IA" : "Modelo")}</span></div></div><div class="text-right"><div class="flex items-center justify-end gap-2">`);
          if (isCompleted(prep)) {
            _push(`<span class="px-2.5 py-1 rounded-lg text-xs font-bold bg-emerald-500/15 border border-emerald-500/30 text-emerald-300"> Concluída </span>`);
          } else {
            _push(`<span class="px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-500/15 border border-amber-500/30 text-amber-300"> Em andamento </span>`);
          }
          if (((_a = verification(prep)) == null ? void 0 : _a.status) === "verified") {
            _push(`<span class="px-2.5 py-1 rounded-lg text-xs font-bold bg-emerald-500/15 border border-emerald-500/30 text-emerald-300">${ssrInterpolate(verification(prep).score.toFixed(1))}/10 </span>`);
          } else if (((_b = verification(prep)) == null ? void 0 : _b.status) === "outdated") {
            _push(`<span class="px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-500/15 border border-amber-500/30 text-amber-300"> Desatualizada </span>`);
          } else if (["running", "pending"].includes((_c = verification(prep)) == null ? void 0 : _c.status)) {
            _push(`<span class="px-2.5 py-1 rounded-lg text-xs font-bold bg-indigo-500/15 border border-indigo-500/30 text-indigo-300"> Verificando... </span>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div><div class="text-3xl font-bold font-outfit text-indigo-400 mt-1">${ssrInterpolate(progressPercent(prep))}%</div><div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Respondidas</div></div></div><div class="mt-6 space-y-3"><div class="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden"><div class="h-full bg-gradient-to-r from-indigo-500 to-emerald-500 transition-all duration-500" style="${ssrRenderStyle({ width: progressPercent(prep) + "%" })}"></div></div><div class="flex items-center justify-between text-sm"><span class="text-slate-400 font-mono">${ssrInterpolate(answeredCount(prep))}/${ssrInterpolate(((_d = prep.progress) == null ? void 0 : _d.total) || 0)} respondidas </span><div class="flex gap-4 text-xs"><span class="text-amber-300">${ssrInterpolate(((_e = prep.progress) == null ? void 0 : _e.practiced) || 0)} praticadas</span><span class="text-emerald-300">${ssrInterpolate(((_f = prep.progress) == null ? void 0 : _f.mastered) || 0)} dominadas</span></div></div></div><div class="mt-6 flex flex-wrap gap-2"><!--[-->`);
          ssrRenderList(prep.tech_stack || [], (tech) => {
            _push(`<div class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">${ssrInterpolate(tech)}</div>`);
          });
          _push(`<!--]--></div><div class="mt-8 flex items-center gap-3 flex-wrap">`);
          _push(ssrRenderComponent(_component_router_link, {
            to: `/interview-prep/${prep.id}`,
            class: "inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(Play), { class: "w-4 h-4" }, null, _parent2, _scopeId));
                _push2(` Retomar Preparação `);
              } else {
                return [
                  createVNode(unref(Play), { class: "w-4 h-4" }),
                  createTextVNode(" Retomar Preparação ")
                ];
              }
            }),
            _: 2
          }, _parent));
          if (canVerify(prep)) {
            _push(`<button${ssrIncludeBooleanAttr(isVerifying(prep)) ? " disabled" : ""} class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-slate-800 border border-white/10 text-slate-200 text-sm font-semibold hover:bg-slate-700 transition-all disabled:opacity-40 disabled:cursor-wait">`);
            if (isVerifying(prep)) {
              _push(`<span class="w-4 h-4 border-2 border-slate-400/30 border-t-slate-200 rounded-full animate-spin"></span>`);
            } else if (verification(prep)) {
              _push(ssrRenderComponent(unref(RefreshCw), { class: "w-4 h-4" }, null, _parent));
            } else {
              _push(ssrRenderComponent(unref(Sparkles), { class: "w-4 h-4" }, null, _parent));
            }
            _push(` ${ssrInterpolate(isVerifying(prep) ? "Corrigindo respostas..." : verifyLabel())}</button>`);
          } else {
            _push(`<!---->`);
          }
          _push(ssrRenderComponent(_component_router_link, {
            to: "/applied",
            class: "inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 transition-all"
          }, {
            default: withCtx((_, _push2, _parent2, _scopeId) => {
              if (_push2) {
                _push2(ssrRenderComponent(unref(Briefcase), { class: "w-4 h-4" }, null, _parent2, _scopeId));
                _push2(` Ver candidatadas `);
              } else {
                return [
                  createVNode(unref(Briefcase), { class: "w-4 h-4" }),
                  createTextVNode(" Ver candidatadas ")
                ];
              }
            }),
            _: 2
          }, _parent));
          _push(`</div></div>`);
        });
        _push(`<!--]--></div>`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$h = _sfc_main$h.setup;
_sfc_main$h.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/InterviewPreparations.vue");
  return _sfc_setup$h ? _sfc_setup$h(props, ctx) : void 0;
};
const _sfc_main$g = {
  __name: "SpontaneousPrep",
  __ssrInlineRender: true,
  setup(__props) {
    useRouter();
    const form = ref({ title: "", company: "", description: "", seed: "" });
    const isLoading = ref(false);
    const error = ref("");
    return (_ctx, _push, _parent, _attrs) => {
      const _component_router_link = resolveComponent("router-link");
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-8" }, _attrs))}><div>`);
      _push(ssrRenderComponent(_component_router_link, {
        to: "/interview-prep",
        class: "inline-flex items-center gap-1.5 text-slate-400 hover:text-slate-200 text-sm mb-2 transition-colors"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(ArrowLeft), { class: "w-4 h-4" }, null, _parent2, _scopeId));
            _push2(` Preparações `);
          } else {
            return [
              createVNode(unref(ArrowLeft), { class: "w-4 h-4" }),
              createTextVNode(" Preparações ")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`<h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Preparação Avulsa</h1><p class="text-slate-400">Cole o conteúdo de uma vaga e gere uma preparação salva com 15 perguntas de entrevista — sem criar uma vaga candidatada.</p></div><div class="glass-card p-8"><div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6"><div><label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Cargo</label><input${ssrRenderAttr("value", form.value.title)} type="text" placeholder="Ex.: Backend Developer" class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors"></div><div><label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Empresa</label><input${ssrRenderAttr("value", form.value.company)} type="text" placeholder="Ex.: Tech Corp" class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors"></div></div><div class="mb-6"><label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2"> Conteúdo da vaga <span class="text-rose-400">*</span></label><textarea rows="6" placeholder="Cole aqui a descrição completa da vaga. Quanto mais conteúdo, melhores as perguntas." class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-3 text-slate-200 text-sm resize-none outline-none focus:border-indigo-500/50 transition-colors">${ssrInterpolate(form.value.description)}</textarea>`);
      if (error.value) {
        _push(`<p class="mt-2 text-sm text-rose-400 font-semibold">${ssrInterpolate(error.value)}</p>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div><div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"><div class="w-full sm:w-auto"><label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Semente (opcional)</label><input${ssrRenderAttr("value", form.value.seed)} type="text" placeholder="Use um valor fixo para a mesma ordem de perguntas" class="w-full sm:w-96 bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors"></div><button${ssrIncludeBooleanAttr(isLoading.value) ? " disabled" : ""} class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed">`);
      if (isLoading.value) {
        _push(`<span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>`);
      } else {
        _push(ssrRenderComponent(unref(Sparkles), { class: "w-4 h-4" }, null, _parent));
      }
      _push(` ${ssrInterpolate(isLoading.value ? "Salvando preparação..." : "Gerar e salvar")}</button></div></div></div>`);
    };
  }
};
const _sfc_setup$g = _sfc_main$g.setup;
_sfc_main$g.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/SpontaneousPrep.vue");
  return _sfc_setup$g ? _sfc_setup$g(props, ctx) : void 0;
};
const _sfc_main$f = {
  __name: "Card",
  __ssrInlineRender: true,
  props: {
    title: { type: String, required: true },
    subtitle: { type: String, default: "" }
  },
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "glass-card p-8" }, _attrs))}><div class="flex items-center justify-between mb-6"><div><h2 class="text-xl font-bold text-white font-outfit">${ssrInterpolate(__props.title)}</h2>`);
      if (__props.subtitle) {
        _push(`<p class="text-sm text-slate-500 mt-1">${ssrInterpolate(__props.subtitle)}</p>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
      ssrRenderSlot(_ctx.$slots, "icon", {}, null, _push, _parent);
      _push(`</div>`);
      ssrRenderSlot(_ctx.$slots, "default", {}, null, _push, _parent);
      _push(`</div>`);
    };
  }
};
const _sfc_setup$f = _sfc_main$f.setup;
_sfc_main$f.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/Card.vue");
  return _sfc_setup$f ? _sfc_setup$f(props, ctx) : void 0;
};
const _sfc_main$e = {
  __name: "EmptyState",
  __ssrInlineRender: true,
  props: {
    data: { type: [Object, Array, null], default: null },
    loading: { type: Boolean, default: false },
    error: { type: Boolean, default: false }
  },
  setup(__props) {
    const props = __props;
    const state = computed(() => {
      var _a;
      return ((_a = props.data) == null ? void 0 : _a.empty_state) || null;
    });
    const empty = computed(() => {
      var _a, _b;
      return !!((_b = (_a = props.data) == null ? void 0 : _a.empty_state) == null ? void 0 : _b.empty);
    });
    return (_ctx, _push, _parent, _attrs) => {
      var _a, _b;
      if (empty.value) {
        _push(`<div${ssrRenderAttrs(mergeProps({ class: "text-center py-12" }, _attrs))}><div class="w-12 h-12 mx-auto mb-4 bg-slate-700/40 rounded-2xl flex items-center justify-center text-slate-400">`);
        _push(ssrRenderComponent(unref(Inbox), { class: "w-6 h-6" }, null, _parent));
        _push(`</div><div class="text-slate-300 font-semibold">${ssrInterpolate((_a = state.value) == null ? void 0 : _a.message)}</div><div class="text-sm text-slate-500 mt-1 max-w-sm mx-auto">${ssrInterpolate((_b = state.value) == null ? void 0 : _b.guidance)}</div></div>`);
      } else if (__props.loading) {
        _push(`<div${ssrRenderAttrs(mergeProps({ class: "flex items-center justify-center h-48" }, _attrs))}><div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div></div>`);
      } else if (__props.error) {
        _push(`<div${ssrRenderAttrs(mergeProps({ class: "text-center py-12 text-red-400 text-sm" }, _attrs))}> Não foi possível carregar os dados de analytics. </div>`);
      } else {
        ssrRenderSlot(_ctx.$slots, "default", {}, null, _push, _parent);
      }
    };
  }
};
const _sfc_setup$e = _sfc_main$e.setup;
_sfc_main$e.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/EmptyState.vue");
  return _sfc_setup$e ? _sfc_setup$e(props, ctx) : void 0;
};
const _sfc_main$d = {
  __name: "Kpi",
  __ssrInlineRender: true,
  props: {
    label: { type: String, required: true },
    value: { type: [Number, String], default: 0 },
    suffix: { type: String, default: "" }
  },
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "bg-slate-900/60 border border-white/5 rounded-xl p-5" }, _attrs))}><div class="text-xs text-slate-500 uppercase tracking-widest mb-2">${ssrInterpolate(__props.label)}</div><div class="text-3xl font-bold text-white font-outfit">${ssrInterpolate(__props.value)}${ssrInterpolate(__props.suffix)}</div></div>`);
    };
  }
};
const _sfc_setup$d = _sfc_main$d.setup;
_sfc_main$d.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/Kpi.vue");
  return _sfc_setup$d ? _sfc_setup$d(props, ctx) : void 0;
};
Chart.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement, LineElement, PointElement);
const PALETTE = ["#6366f1", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#06b6d4", "#ec4899", "#84cc16", "#f97316", "#3b82f6"];
const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: { y: { beginAtZero: true } }
};
const pieOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: "right" } }
};
const _sfc_main$c = {
  __name: "MatchingWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsMatching();
    const summary = computed(() => {
      var _a;
      return ((_a = data.value) == null ? void 0 : _a.summary) || {};
    });
    const titles = computed(() => {
      var _a;
      return ((_a = data.value) == null ? void 0 : _a.top_job_titles) || [];
    });
    const bucketChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.buckets) || []).map((b) => b.bucket),
        datasets: [{
          label: "Matches",
          data: (((_b = data.value) == null ? void 0 : _b.buckets) || []).map((b) => b.count),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    function pct(v) {
      return Math.round((Number(v) || 0) * 100);
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Efetividade do Matching",
        subtitle: "Distribuição de similaridade, média, mediana e conversão"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Target), { class: "w-6 h-6 text-indigo-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Target), { class: "w-6 h-6 text-indigo-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Total de Matches",
                    value: (_a = summary.value) == null ? void 0 : _a.total_matches
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Score Médio",
                    value: pct((_b = summary.value) == null ? void 0 : _b.average_score),
                    suffix: "%"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Score Mediano",
                    value: pct((_c = summary.value) == null ? void 0 : _c.median_score),
                    suffix: "%"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Taxa de Conversão",
                    value: pct((_d = summary.value) == null ? void 0 : _d.conversion_rate),
                    suffix: "%"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Distribuição de Scores</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: bucketChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Melhores Títulos</h3>`);
                  if (titles.value.length) {
                    _push3(`<div class="space-y-3"${_scopeId2}><!--[-->`);
                    ssrRenderList(titles.value, (t) => {
                      _push3(`<div class="flex items-center gap-4"${_scopeId2}><span class="flex-1 text-sm text-slate-300 truncate"${_scopeId2}>${ssrInterpolate(t.title)}</span><div class="w-32 bg-slate-700 rounded-full h-1.5"${_scopeId2}><div class="bg-indigo-500 h-1.5 rounded-full" style="${ssrRenderStyle({ width: t.success_rate + "%" })}"${_scopeId2}></div></div><span class="text-xs text-slate-400 w-10 text-right"${_scopeId2}>${ssrInterpolate(t.success_rate.toFixed(1))}%</span></div>`);
                    });
                    _push3(`<!--]--></div>`);
                  } else {
                    _push3(`<div class="text-slate-500 text-sm py-8 text-center"${_scopeId2}>Sem candidaturas registradas ainda.</div>`);
                  }
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Total de Matches",
                        value: (_e = summary.value) == null ? void 0 : _e.total_matches
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: pct((_f = summary.value) == null ? void 0 : _f.average_score),
                        suffix: "%"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Mediano",
                        value: pct((_g = summary.value) == null ? void 0 : _g.median_score),
                        suffix: "%"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Taxa de Conversão",
                        value: pct((_h = summary.value) == null ? void 0 : _h.conversion_rate),
                        suffix: "%"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Distribuição de Scores"),
                        createVNode(unref(Bar), {
                          data: bucketChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", null, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Melhores Títulos"),
                        titles.value.length ? (openBlock(), createBlock("div", {
                          key: 0,
                          class: "space-y-3"
                        }, [
                          (openBlock(true), createBlock(Fragment, null, renderList(titles.value, (t) => {
                            return openBlock(), createBlock("div", {
                              key: t.title,
                              class: "flex items-center gap-4"
                            }, [
                              createVNode("span", { class: "flex-1 text-sm text-slate-300 truncate" }, toDisplayString(t.title), 1),
                              createVNode("div", { class: "w-32 bg-slate-700 rounded-full h-1.5" }, [
                                createVNode("div", {
                                  class: "bg-indigo-500 h-1.5 rounded-full",
                                  style: { width: t.success_rate + "%" }
                                }, null, 4)
                              ]),
                              createVNode("span", { class: "text-xs text-slate-400 w-10 text-right" }, toDisplayString(t.success_rate.toFixed(1)) + "%", 1)
                            ]);
                          }), 128))
                        ])) : (openBlock(), createBlock("div", {
                          key: 1,
                          class: "text-slate-500 text-sm py-8 text-center"
                        }, "Sem candidaturas registradas ainda."))
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Total de Matches",
                        value: (_a = summary.value) == null ? void 0 : _a.total_matches
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: pct((_b = summary.value) == null ? void 0 : _b.average_score),
                        suffix: "%"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Mediano",
                        value: pct((_c = summary.value) == null ? void 0 : _c.median_score),
                        suffix: "%"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Taxa de Conversão",
                        value: pct((_d = summary.value) == null ? void 0 : _d.conversion_rate),
                        suffix: "%"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Distribuição de Scores"),
                        createVNode(unref(Bar), {
                          data: bucketChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", null, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Melhores Títulos"),
                        titles.value.length ? (openBlock(), createBlock("div", {
                          key: 0,
                          class: "space-y-3"
                        }, [
                          (openBlock(true), createBlock(Fragment, null, renderList(titles.value, (t) => {
                            return openBlock(), createBlock("div", {
                              key: t.title,
                              class: "flex items-center gap-4"
                            }, [
                              createVNode("span", { class: "flex-1 text-sm text-slate-300 truncate" }, toDisplayString(t.title), 1),
                              createVNode("div", { class: "w-32 bg-slate-700 rounded-full h-1.5" }, [
                                createVNode("div", {
                                  class: "bg-indigo-500 h-1.5 rounded-full",
                                  style: { width: t.success_rate + "%" }
                                }, null, 4)
                              ]),
                              createVNode("span", { class: "text-xs text-slate-400 w-10 text-right" }, toDisplayString(t.success_rate.toFixed(1)) + "%", 1)
                            ]);
                          }), 128))
                        ])) : (openBlock(), createBlock("div", {
                          key: 1,
                          class: "text-slate-500 text-sm py-8 text-center"
                        }, "Sem candidaturas registradas ainda."))
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$c = _sfc_main$c.setup;
_sfc_main$c.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/MatchingWidget.vue");
  return _sfc_setup$c ? _sfc_setup$c(props, ctx) : void 0;
};
const _sfc_main$b = {
  __name: "FacetList",
  __ssrInlineRender: true,
  props: {
    title: { type: String, required: true },
    items: { type: Array, default: () => [] },
    color: { type: String, default: "bg-slate-500" },
    labelKey: { type: String, default: "skill" }
  },
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      var _a;
      _push(`<div${ssrRenderAttrs(_attrs)}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3">${ssrInterpolate(__props.title)}</h3><ul class="space-y-3"><!--[-->`);
      ssrRenderList(__props.items, (item, i) => {
        _push(`<li class="flex items-center gap-3"><span class="${ssrRenderClass([__props.color, "w-1.5 h-1.5 rounded-full"])}"></span><span class="flex-1 text-sm text-slate-300 truncate">${ssrInterpolate(item[__props.labelKey] || item.skill || item.name)}</span><span class="text-xs font-bold text-slate-400">${ssrInterpolate(item.count)}</span></li>`);
      });
      _push(`<!--]--></ul>`);
      if (!((_a = __props.items) == null ? void 0 : _a.length)) {
        _push(`<div class="text-slate-500 text-sm py-6 text-center">Sem dados.</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
};
const _sfc_setup$b = _sfc_main$b.setup;
_sfc_main$b.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/FacetList.vue");
  return _sfc_setup$b ? _sfc_setup$b(props, ctx) : void 0;
};
const _sfc_main$a = {
  __name: "ResumeWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsResume();
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Qualidade de Currículos",
        subtitle: "Pontos fortes, fracos e melhorias mais frequentes"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(FileText), { class: "w-6 h-6 text-emerald-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(FileText), { class: "w-6 h-6 text-emerald-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-1 lg:grid-cols-3 gap-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$b, {
                    title: "Pontos Fortes",
                    items: (_a = unref(data)) == null ? void 0 : _a.top_strengths,
                    color: "bg-emerald-500"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$b, {
                    title: "Pontos Fracos",
                    items: (_b = unref(data)) == null ? void 0 : _b.top_weaknesses,
                    color: "bg-red-500"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$b, {
                    title: "Sugestões",
                    items: (_c = unref(data)) == null ? void 0 : _c.top_suggestions,
                    color: "bg-amber-500",
                    "label-key": "suggestion"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-3 gap-8" }, [
                      createVNode(_sfc_main$b, {
                        title: "Pontos Fortes",
                        items: (_d = unref(data)) == null ? void 0 : _d.top_strengths,
                        color: "bg-emerald-500"
                      }, null, 8, ["items"]),
                      createVNode(_sfc_main$b, {
                        title: "Pontos Fracos",
                        items: (_e = unref(data)) == null ? void 0 : _e.top_weaknesses,
                        color: "bg-red-500"
                      }, null, 8, ["items"]),
                      createVNode(_sfc_main$b, {
                        title: "Sugestões",
                        items: (_f = unref(data)) == null ? void 0 : _f.top_suggestions,
                        color: "bg-amber-500",
                        "label-key": "suggestion"
                      }, null, 8, ["items"])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c;
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-3 gap-8" }, [
                      createVNode(_sfc_main$b, {
                        title: "Pontos Fortes",
                        items: (_a = unref(data)) == null ? void 0 : _a.top_strengths,
                        color: "bg-emerald-500"
                      }, null, 8, ["items"]),
                      createVNode(_sfc_main$b, {
                        title: "Pontos Fracos",
                        items: (_b = unref(data)) == null ? void 0 : _b.top_weaknesses,
                        color: "bg-red-500"
                      }, null, 8, ["items"]),
                      createVNode(_sfc_main$b, {
                        title: "Sugestões",
                        items: (_c = unref(data)) == null ? void 0 : _c.top_suggestions,
                        color: "bg-amber-500",
                        "label-key": "suggestion"
                      }, null, 8, ["items"])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$a = _sfc_main$a.setup;
_sfc_main$a.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/ResumeWidget.vue");
  return _sfc_setup$a ? _sfc_setup$a(props, ctx) : void 0;
};
const _sfc_main$9 = {
  __name: "JobTrendsWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsJobTrends();
    const topUsers = computed(() => {
      var _a;
      return (((_a = data.value) == null ? void 0 : _a.top_users) || []).map((u) => ({ name: u.name, count: u.job_count }));
    });
    const statusChart = computed(() => {
      var _a, _b;
      return {
        labels: Object.keys(((_a = data.value) == null ? void 0 : _a.jobs_by_status) || {}),
        datasets: [{
          label: "Vagas",
          data: Object.values(((_b = data.value) == null ? void 0 : _b.jobs_by_status) || {}),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Tendências de Vagas",
        subtitle: "Status, principais usuários e volume de postagens"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Briefcase), { class: "w-6 h-6 text-amber-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Briefcase), { class: "w-6 h-6 text-amber-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                if (_push3) {
                  _push3(`<div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Vagas por Status</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: statusChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Principais Usuários</h3>`);
                  _push3(ssrRenderComponent(_sfc_main$b, {
                    items: topUsers.value,
                    color: "bg-amber-500"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Vagas por Status"),
                        createVNode(unref(Bar), {
                          data: statusChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", null, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Principais Usuários"),
                        createVNode(_sfc_main$b, {
                          items: topUsers.value,
                          color: "bg-amber-500"
                        }, null, 8, ["items"])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => [
                  createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                    createVNode("div", { class: "h-[260px]" }, [
                      createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Vagas por Status"),
                      createVNode(unref(Bar), {
                        data: statusChart.value,
                        options: unref(barOptions)
                      }, null, 8, ["data", "options"])
                    ]),
                    createVNode("div", null, [
                      createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Principais Usuários"),
                      createVNode(_sfc_main$b, {
                        items: topUsers.value,
                        color: "bg-amber-500"
                      }, null, 8, ["items"])
                    ])
                  ])
                ]),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$9 = _sfc_main$9.setup;
_sfc_main$9.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/JobTrendsWidget.vue");
  return _sfc_setup$9 ? _sfc_setup$9(props, ctx) : void 0;
};
const _sfc_main$8 = {
  __name: "UserMetricsWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data: me, isLoading: meLoading } = useMe();
    const { user } = useAuth();
    const isAdmin = computed(() => {
      var _a;
      return ((_a = user.value) == null ? void 0 : _a.role) === "admin";
    });
    const userId = computed(() => {
      var _a, _b;
      return isAdmin.value ? 0 : (_b = (_a = me.value) == null ? void 0 : _a.user) == null ? void 0 : _b.id;
    });
    const { data, isLoading: loading, isError: error } = useAnalyticsUserMetrics(userId);
    const metrics = computed(() => {
      var _a;
      return (_a = data.value) == null ? void 0 : _a.metrics;
    });
    function pct(v) {
      return Math.round((Number(v) || 0) * 100);
    }
    const FEATURES = { matching: "Matching", ai_analysis: "Análise de IA", interview_prep: "Entrevista", skill_gap: "Skill Gap" };
    function featureLabel(key) {
      return FEATURES[key] || key;
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Métricas do Usuário",
        subtitle: "Plano, assinatura e conversão"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(User), { class: "w-6 h-6 text-purple-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(User), { class: "w-6 h-6 text-purple-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Vagas",
                    value: (_a = metrics.value) == null ? void 0 : _a.jobs_posted
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Currículos",
                    value: (_b = metrics.value) == null ? void 0 : _b.resumes_uploaded
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Matches",
                    value: (_c = metrics.value) == null ? void 0 : _c.matches_made
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Conversão",
                    value: pct((_d = metrics.value) == null ? void 0 : _d.conversion_rate),
                    suffix: "%"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-center"${_scopeId2}><!--[-->`);
                  ssrRenderList(((_e = metrics.value) == null ? void 0 : _e.feature_utilization) || {}, (active, key) => {
                    _push3(`<div class="bg-slate-900/60 border border-white/5 rounded-xl p-4"${_scopeId2}><div class="text-xs text-slate-500 uppercase tracking-widest mb-2"${_scopeId2}>${ssrInterpolate(featureLabel(key))}</div><div class="${ssrRenderClass([active ? "text-emerald-400" : "text-slate-600", "font-bold"])}"${_scopeId2}>${ssrInterpolate(active ? "Ativo" : "Inativo")}</div></div>`);
                  });
                  _push3(`<!--]--></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Vagas",
                        value: (_f = metrics.value) == null ? void 0 : _f.jobs_posted
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Currículos",
                        value: (_g = metrics.value) == null ? void 0 : _g.resumes_uploaded
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Matches",
                        value: (_h = metrics.value) == null ? void 0 : _h.matches_made
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Conversão",
                        value: pct((_i = metrics.value) == null ? void 0 : _i.conversion_rate),
                        suffix: "%"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-2 md:grid-cols-4 gap-4 text-center" }, [
                      (openBlock(true), createBlock(Fragment, null, renderList(((_j = metrics.value) == null ? void 0 : _j.feature_utilization) || {}, (active, key) => {
                        return openBlock(), createBlock("div", {
                          key,
                          class: "bg-slate-900/60 border border-white/5 rounded-xl p-4"
                        }, [
                          createVNode("div", { class: "text-xs text-slate-500 uppercase tracking-widest mb-2" }, toDisplayString(featureLabel(key)), 1),
                          createVNode("div", {
                            class: [active ? "text-emerald-400" : "text-slate-600", "font-bold"]
                          }, toDisplayString(active ? "Ativo" : "Inativo"), 3)
                        ]);
                      }), 128))
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d, _e;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Vagas",
                        value: (_a = metrics.value) == null ? void 0 : _a.jobs_posted
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Currículos",
                        value: (_b = metrics.value) == null ? void 0 : _b.resumes_uploaded
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Matches",
                        value: (_c = metrics.value) == null ? void 0 : _c.matches_made
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Conversão",
                        value: pct((_d = metrics.value) == null ? void 0 : _d.conversion_rate),
                        suffix: "%"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-2 md:grid-cols-4 gap-4 text-center" }, [
                      (openBlock(true), createBlock(Fragment, null, renderList(((_e = metrics.value) == null ? void 0 : _e.feature_utilization) || {}, (active, key) => {
                        return openBlock(), createBlock("div", {
                          key,
                          class: "bg-slate-900/60 border border-white/5 rounded-xl p-4"
                        }, [
                          createVNode("div", { class: "text-xs text-slate-500 uppercase tracking-widest mb-2" }, toDisplayString(featureLabel(key)), 1),
                          createVNode("div", {
                            class: [active ? "text-emerald-400" : "text-slate-600", "font-bold"]
                          }, toDisplayString(active ? "Ativo" : "Inativo"), 3)
                        ]);
                      }), 128))
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$8 = _sfc_main$8.setup;
_sfc_main$8.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/UserMetricsWidget.vue");
  return _sfc_setup$8 ? _sfc_setup$8(props, ctx) : void 0;
};
const _sfc_main$7 = {
  __name: "InterviewPrepWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsInterviewPrep();
    const categoryChart = computed(() => {
      var _a, _b;
      return {
        labels: Object.keys(((_a = data.value) == null ? void 0 : _a.category_distribution) || {}),
        datasets: [{ data: Object.values(((_b = data.value) == null ? void 0 : _b.category_distribution) || {}), backgroundColor: PALETTE }]
      };
    });
    const sourceChart = computed(() => {
      var _a, _b;
      return {
        labels: Object.keys(((_a = data.value) == null ? void 0 : _a.source_distribution) || {}),
        datasets: [{ label: "Preparações", data: Object.values(((_b = data.value) == null ? void 0 : _b.source_distribution) || {}), backgroundColor: PALETTE, borderRadius: 4 }]
      };
    });
    const categoryCount = computed(() => {
      var _a;
      return Object.keys(((_a = data.value) == null ? void 0 : _a.category_distribution) || {}).length;
    });
    function trendLabel(t) {
      return { improving: "Melhorando", stable: "Estável", declining: "Caindo" }[t] || t || "—";
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Preparação para Entrevistas",
        subtitle: "Categorias de perguntas, fontes e score de verificação"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(GraduationCap), { class: "w-6 h-6 text-pink-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(GraduationCap), { class: "w-6 h-6 text-pink-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Preparações",
                    value: (_a = unref(data)) == null ? void 0 : _a.total_preparations
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Score Médio",
                    value: (_d = (_c = (_b = unref(data)) == null ? void 0 : _b.average_verification_score) == null ? void 0 : _c.toFixed) == null ? void 0 : _d.call(_c, 1)
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Tendência",
                    value: trendLabel((_e = unref(data)) == null ? void 0 : _e.verification_trend)
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Categorias",
                    value: categoryCount.value
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Por Categoria</h3>`);
                  _push3(ssrRenderComponent(unref(Doughnut), {
                    data: categoryChart.value,
                    options: unref(pieOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Por Fonte</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: sourceChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Preparações",
                        value: (_f = unref(data)) == null ? void 0 : _f.total_preparations
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: (_i = (_h = (_g = unref(data)) == null ? void 0 : _g.average_verification_score) == null ? void 0 : _h.toFixed) == null ? void 0 : _i.call(_h, 1)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tendência",
                        value: trendLabel((_j = unref(data)) == null ? void 0 : _j.verification_trend)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Categorias",
                        value: categoryCount.value
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Por Categoria"),
                        createVNode(unref(Doughnut), {
                          data: categoryChart.value,
                          options: unref(pieOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Por Fonte"),
                        createVNode(unref(Bar), {
                          data: sourceChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d, _e;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Preparações",
                        value: (_a = unref(data)) == null ? void 0 : _a.total_preparations
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: (_d = (_c = (_b = unref(data)) == null ? void 0 : _b.average_verification_score) == null ? void 0 : _c.toFixed) == null ? void 0 : _d.call(_c, 1)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tendência",
                        value: trendLabel((_e = unref(data)) == null ? void 0 : _e.verification_trend)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Categorias",
                        value: categoryCount.value
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Por Categoria"),
                        createVNode(unref(Doughnut), {
                          data: categoryChart.value,
                          options: unref(pieOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Por Fonte"),
                        createVNode(unref(Bar), {
                          data: sourceChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$7 = _sfc_main$7.setup;
_sfc_main$7.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/InterviewPrepWidget.vue");
  return _sfc_setup$7 ? _sfc_setup$7(props, ctx) : void 0;
};
const _sfc_main$6 = {
  __name: "AIQualityWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsAIQuality();
    const bucketChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.score_buckets) || []).map((b) => b.bucket),
        datasets: [{
          label: "Análises",
          data: (((_b = data.value) == null ? void 0 : _b.score_buckets) || []).map((b) => b.count),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    const sourceList = computed(
      () => {
        var _a;
        return (((_a = data.value) == null ? void 0 : _a.success_rate_by_source) || []).map((s) => {
          var _a2, _b;
          return { source: s.source, count: s.completed, name: `${(_b = (_a2 = s.success_rate) == null ? void 0 : _a2.toFixed) == null ? void 0 : _b.call(_a2, 1)}%` };
        });
      }
    );
    function trendLabel(t) {
      return { improving: "Melhorando", stable: "Estável", declining: "Caindo" }[t] || t || "—";
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Qualidade de Análises de IA",
        subtitle: "Distribuição de score, status e sucesso por fonte"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Sparkles), { class: "w-6 h-6 text-cyan-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Sparkles), { class: "w-6 h-6 text-cyan-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Análises",
                    value: (_a = unref(data)) == null ? void 0 : _a.total_analyses
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Score Médio",
                    value: (_d = (_c = (_b = unref(data)) == null ? void 0 : _b.average_score) == null ? void 0 : _c.toFixed) == null ? void 0 : _d.call(_c, 1)
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Tempo Médio",
                    value: (_e = unref(data)) == null ? void 0 : _e.average_processing_seconds,
                    suffix: "s"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Tendência",
                    value: trendLabel((_f = unref(data)) == null ? void 0 : _f.verification_trend)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Score por Bucket</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: bucketChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Sucesso por Fonte</h3>`);
                  _push3(ssrRenderComponent(_sfc_main$b, {
                    items: sourceList.value,
                    color: "bg-cyan-500",
                    "label-key": "source"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Análises",
                        value: (_g = unref(data)) == null ? void 0 : _g.total_analyses
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: (_j = (_i = (_h = unref(data)) == null ? void 0 : _h.average_score) == null ? void 0 : _i.toFixed) == null ? void 0 : _j.call(_i, 1)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tempo Médio",
                        value: (_k = unref(data)) == null ? void 0 : _k.average_processing_seconds,
                        suffix: "s"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tendência",
                        value: trendLabel((_l = unref(data)) == null ? void 0 : _l.verification_trend)
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Score por Bucket"),
                        createVNode(unref(Bar), {
                          data: bucketChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", null, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Sucesso por Fonte"),
                        createVNode(_sfc_main$b, {
                          items: sourceList.value,
                          color: "bg-cyan-500",
                          "label-key": "source"
                        }, null, 8, ["items"])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d, _e, _f;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Análises",
                        value: (_a = unref(data)) == null ? void 0 : _a.total_analyses
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Score Médio",
                        value: (_d = (_c = (_b = unref(data)) == null ? void 0 : _b.average_score) == null ? void 0 : _c.toFixed) == null ? void 0 : _d.call(_c, 1)
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tempo Médio",
                        value: (_e = unref(data)) == null ? void 0 : _e.average_processing_seconds,
                        suffix: "s"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Tendência",
                        value: trendLabel((_f = unref(data)) == null ? void 0 : _f.verification_trend)
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Score por Bucket"),
                        createVNode(unref(Bar), {
                          data: bucketChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", null, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Sucesso por Fonte"),
                        createVNode(_sfc_main$b, {
                          items: sourceList.value,
                          color: "bg-cyan-500",
                          "label-key": "source"
                        }, null, 8, ["items"])
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$6 = _sfc_main$6.setup;
_sfc_main$6.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/AIQualityWidget.vue");
  return _sfc_setup$6 ? _sfc_setup$6(props, ctx) : void 0;
};
const _sfc_main$5 = {
  __name: "TimeToHireWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsTimeToHire();
    const stageChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.stage_timings) || []).map((s) => s.stage),
        datasets: [{
          label: "Dias",
          data: (((_b = data.value) == null ? void 0 : _b.stage_timings) || []).map((s) => s.median_days),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    function days(v) {
      return (Number(v) || 0).toFixed(1);
    }
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Tempo até a Contratação",
        subtitle: "Duração média e mediana por estágio do pipeline"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Timer), { class: "w-6 h-6 text-orange-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Timer), { class: "w-6 h-6 text-orange-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-6 gap-4 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Match (média)",
                    value: days((_a = unref(data)) == null ? void 0 : _a.average_time_to_match_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Análise (média)",
                    value: days((_b = unref(data)) == null ? void 0 : _b.average_time_to_analysis_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Apply (média)",
                    value: days((_c = unref(data)) == null ? void 0 : _c.average_time_to_apply_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Match (mediana)",
                    value: days((_d = unref(data)) == null ? void 0 : _d.median_time_to_match_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Análise (mediana)",
                    value: days((_e = unref(data)) == null ? void 0 : _e.median_time_to_analysis_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Apply (mediana)",
                    value: days((_f = unref(data)) == null ? void 0 : _f.median_time_to_apply_days),
                    suffix: "d"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="h-[280px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Tempo Mediano por Estágio</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: stageChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-6 gap-4 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Match (média)",
                        value: days((_g = unref(data)) == null ? void 0 : _g.average_time_to_match_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Análise (média)",
                        value: days((_h = unref(data)) == null ? void 0 : _h.average_time_to_analysis_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Apply (média)",
                        value: days((_i = unref(data)) == null ? void 0 : _i.average_time_to_apply_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Match (mediana)",
                        value: days((_j = unref(data)) == null ? void 0 : _j.median_time_to_match_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Análise (mediana)",
                        value: days((_k = unref(data)) == null ? void 0 : _k.median_time_to_analysis_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Apply (mediana)",
                        value: days((_l = unref(data)) == null ? void 0 : _l.median_time_to_apply_days),
                        suffix: "d"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "h-[280px]" }, [
                      createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Tempo Mediano por Estágio"),
                      createVNode(unref(Bar), {
                        data: stageChart.value,
                        options: unref(barOptions)
                      }, null, 8, ["data", "options"])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d, _e, _f;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-6 gap-4 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Match (média)",
                        value: days((_a = unref(data)) == null ? void 0 : _a.average_time_to_match_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Análise (média)",
                        value: days((_b = unref(data)) == null ? void 0 : _b.average_time_to_analysis_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Apply (média)",
                        value: days((_c = unref(data)) == null ? void 0 : _c.average_time_to_apply_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Match (mediana)",
                        value: days((_d = unref(data)) == null ? void 0 : _d.median_time_to_match_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Análise (mediana)",
                        value: days((_e = unref(data)) == null ? void 0 : _e.median_time_to_analysis_days),
                        suffix: "d"
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Apply (mediana)",
                        value: days((_f = unref(data)) == null ? void 0 : _f.median_time_to_apply_days),
                        suffix: "d"
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "h-[280px]" }, [
                      createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Tempo Mediano por Estágio"),
                      createVNode(unref(Bar), {
                        data: stageChart.value,
                        options: unref(barOptions)
                      }, null, 8, ["data", "options"])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$5 = _sfc_main$5.setup;
_sfc_main$5.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/TimeToHireWidget.vue");
  return _sfc_setup$5 ? _sfc_setup$5(props, ctx) : void 0;
};
const _sfc_main$4 = {
  __name: "ChipList",
  __ssrInlineRender: true,
  props: {
    title: { type: String, required: true },
    items: { type: Array, default: () => [] },
    color: { type: String, default: "bg-slate-500" }
  },
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      var _a;
      _push(`<div${ssrRenderAttrs(_attrs)}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3">${ssrInterpolate(__props.title)}</h3><div class="flex flex-wrap gap-2"><!--[-->`);
      ssrRenderList(__props.items, (item) => {
        _push(`<span class="px-3 py-1 bg-slate-900/60 border border-white/5 rounded-full text-xs text-slate-300">${ssrInterpolate(item)}</span>`);
      });
      _push(`<!--]-->`);
      if (!((_a = __props.items) == null ? void 0 : _a.length)) {
        _push(`<span class="text-slate-500 text-sm">Sem dados.</span>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div></div>`);
    };
  }
};
const _sfc_setup$4 = _sfc_main$4.setup;
_sfc_main$4.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/ChipList.vue");
  return _sfc_setup$4 ? _sfc_setup$4(props, ctx) : void 0;
};
const _sfc_main$3 = {
  __name: "SkillGapWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsSkillsGap();
    const categoryChart = computed(() => {
      var _a, _b;
      return {
        labels: Object.keys(((_a = data.value) == null ? void 0 : _a.match_rate_by_category) || {}),
        datasets: [{
          label: "%",
          data: Object.values(((_b = data.value) == null ? void 0 : _b.match_rate_by_category) || {}),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Gap de Skills",
        subtitle: "Sobreposição entre vagas e candidatos por categoria"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Puzzle), { class: "w-6 h-6 text-lime-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Puzzle), { class: "w-6 h-6 text-lime-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[280px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Match Rate por Categoria</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: categoryChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="space-y-6"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$4, {
                    title: "Habilidades em Demanda",
                    items: (_a = unref(data)) == null ? void 0 : _a.in_demand_skills,
                    color: "bg-lime-500"
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$4, {
                    title: "Habilidades Sobrepostas",
                    items: (_b = unref(data)) == null ? void 0 : _b.overlapping_skills,
                    color: "bg-indigo-500"
                  }, null, _parent3, _scopeId2));
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[280px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Match Rate por Categoria"),
                        createVNode(unref(Bar), {
                          data: categoryChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "space-y-6" }, [
                        createVNode(_sfc_main$4, {
                          title: "Habilidades em Demanda",
                          items: (_c = unref(data)) == null ? void 0 : _c.in_demand_skills,
                          color: "bg-lime-500"
                        }, null, 8, ["items"]),
                        createVNode(_sfc_main$4, {
                          title: "Habilidades Sobrepostas",
                          items: (_d = unref(data)) == null ? void 0 : _d.overlapping_skills,
                          color: "bg-indigo-500"
                        }, null, 8, ["items"])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b;
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[280px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Match Rate por Categoria"),
                        createVNode(unref(Bar), {
                          data: categoryChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "space-y-6" }, [
                        createVNode(_sfc_main$4, {
                          title: "Habilidades em Demanda",
                          items: (_a = unref(data)) == null ? void 0 : _a.in_demand_skills,
                          color: "bg-lime-500"
                        }, null, 8, ["items"]),
                        createVNode(_sfc_main$4, {
                          title: "Habilidades Sobrepostas",
                          items: (_b = unref(data)) == null ? void 0 : _b.overlapping_skills,
                          color: "bg-indigo-500"
                        }, null, 8, ["items"])
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$3 = _sfc_main$3.setup;
_sfc_main$3.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/SkillGapWidget.vue");
  return _sfc_setup$3 ? _sfc_setup$3(props, ctx) : void 0;
};
const _sfc_main$2 = {
  __name: "FunnelWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsFunnel();
    const funnelChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.funnel_stages) || []).map((s) => s.stage),
        datasets: [{
          label: "Usuários",
          data: (((_b = data.value) == null ? void 0 : _b.funnel_stages) || []).map((s) => s.count),
          backgroundColor: PALETTE,
          borderRadius: 4
        }]
      };
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Funil de Conversão",
        subtitle: "Estágios do usuário e pontos de abandono"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(Filter), { class: "w-6 h-6 text-rose-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(Filter), { class: "w-6 h-6 text-rose-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[280px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Estágios do Funil</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: funnelChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="space-y-6"${_scopeId2}><div class="bg-slate-900/60 border border-white/5 rounded-xl p-5"${_scopeId2}><div class="text-xs text-slate-500 uppercase tracking-widest mb-2"${_scopeId2}>Total de Usuários</div><div class="text-3xl font-bold text-white font-outfit"${_scopeId2}>${ssrInterpolate((_a = unref(data)) == null ? void 0 : _a.total_users)}</div></div><div${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3"${_scopeId2}>Pontos de Abandono</h3><ul class="space-y-3"${_scopeId2}><!--[-->`);
                  ssrRenderList(((_b = unref(data)) == null ? void 0 : _b.drop_off_points) || [], (stage) => {
                    _push3(`<li class="flex items-center gap-3"${_scopeId2}><span class="w-1.5 h-1.5 rounded-full bg-rose-500"${_scopeId2}></span><span class="text-sm text-slate-300"${_scopeId2}>${ssrInterpolate(stage)}</span></li>`);
                  });
                  _push3(`<!--]--></ul>`);
                  if (!((_d = (_c = unref(data)) == null ? void 0 : _c.drop_off_points) == null ? void 0 : _d.length)) {
                    _push3(`<div class="text-slate-500 text-sm py-4 text-center"${_scopeId2}>Sem pontos de abandono.</div>`);
                  } else {
                    _push3(`<!---->`);
                  }
                  _push3(`</div></div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[280px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Estágios do Funil"),
                        createVNode(unref(Bar), {
                          data: funnelChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "space-y-6" }, [
                        createVNode("div", { class: "bg-slate-900/60 border border-white/5 rounded-xl p-5" }, [
                          createVNode("div", { class: "text-xs text-slate-500 uppercase tracking-widest mb-2" }, "Total de Usuários"),
                          createVNode("div", { class: "text-3xl font-bold text-white font-outfit" }, toDisplayString((_e = unref(data)) == null ? void 0 : _e.total_users), 1)
                        ]),
                        createVNode("div", null, [
                          createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-3" }, "Pontos de Abandono"),
                          createVNode("ul", { class: "space-y-3" }, [
                            (openBlock(true), createBlock(Fragment, null, renderList(((_f = unref(data)) == null ? void 0 : _f.drop_off_points) || [], (stage) => {
                              return openBlock(), createBlock("li", {
                                key: stage,
                                class: "flex items-center gap-3"
                              }, [
                                createVNode("span", { class: "w-1.5 h-1.5 rounded-full bg-rose-500" }),
                                createVNode("span", { class: "text-sm text-slate-300" }, toDisplayString(stage), 1)
                              ]);
                            }), 128))
                          ]),
                          !((_h = (_g = unref(data)) == null ? void 0 : _g.drop_off_points) == null ? void 0 : _h.length) ? (openBlock(), createBlock("div", {
                            key: 0,
                            class: "text-slate-500 text-sm py-4 text-center"
                          }, "Sem pontos de abandono.")) : createCommentVNode("", true)
                        ])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d;
                  return [
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[280px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Estágios do Funil"),
                        createVNode(unref(Bar), {
                          data: funnelChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "space-y-6" }, [
                        createVNode("div", { class: "bg-slate-900/60 border border-white/5 rounded-xl p-5" }, [
                          createVNode("div", { class: "text-xs text-slate-500 uppercase tracking-widest mb-2" }, "Total de Usuários"),
                          createVNode("div", { class: "text-3xl font-bold text-white font-outfit" }, toDisplayString((_a = unref(data)) == null ? void 0 : _a.total_users), 1)
                        ]),
                        createVNode("div", null, [
                          createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-3" }, "Pontos de Abandono"),
                          createVNode("ul", { class: "space-y-3" }, [
                            (openBlock(true), createBlock(Fragment, null, renderList(((_b = unref(data)) == null ? void 0 : _b.drop_off_points) || [], (stage) => {
                              return openBlock(), createBlock("li", {
                                key: stage,
                                class: "flex items-center gap-3"
                              }, [
                                createVNode("span", { class: "w-1.5 h-1.5 rounded-full bg-rose-500" }),
                                createVNode("span", { class: "text-sm text-slate-300" }, toDisplayString(stage), 1)
                              ]);
                            }), 128))
                          ]),
                          !((_d = (_c = unref(data)) == null ? void 0 : _c.drop_off_points) == null ? void 0 : _d.length) ? (openBlock(), createBlock("div", {
                            key: 0,
                            class: "text-slate-500 text-sm py-4 text-center"
                          }, "Sem pontos de abandono.")) : createCommentVNode("", true)
                        ])
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$2 = _sfc_main$2.setup;
_sfc_main$2.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/FunnelWidget.vue");
  return _sfc_setup$2 ? _sfc_setup$2(props, ctx) : void 0;
};
const _sfc_main$1 = {
  __name: "GeographicWidget",
  __ssrInlineRender: true,
  setup(__props) {
    const { data, isLoading: loading, isError: error } = useAnalyticsGeographic();
    const usersChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.users_by_city) || []).map((c) => c.city),
        datasets: [{ label: "Usuários", data: (((_b = data.value) == null ? void 0 : _b.users_by_city) || []).map((c) => c.count), backgroundColor: PALETTE, borderRadius: 4 }]
      };
    });
    const jobsChart = computed(() => {
      var _a, _b;
      return {
        labels: (((_a = data.value) == null ? void 0 : _a.jobs_by_city) || []).map((c) => c.city),
        datasets: [{ label: "Vagas", data: (((_b = data.value) == null ? void 0 : _b.jobs_by_city) || []).map((c) => c.count), backgroundColor: PALETTE, borderRadius: 4 }]
      };
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(ssrRenderComponent(_sfc_main$f, mergeProps({
        title: "Insights Geográficos",
        subtitle: "Distribuição de usuários e vagas, e proporção remoto/presencial"
      }, _attrs), {
        icon: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(unref(MapPin), { class: "w-6 h-6 text-teal-400" }, null, _parent2, _scopeId));
          } else {
            return [
              createVNode(unref(MapPin), { class: "w-6 h-6 text-teal-400" })
            ];
          }
        }),
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_sfc_main$e, {
              data: unref(data),
              loading: unref(loading),
              error: unref(error)
            }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n;
                if (_push3) {
                  _push3(`<div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8"${_scopeId2}>`);
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Cidades",
                    value: (_b = (_a = unref(data)) == null ? void 0 : _a.users_by_city) == null ? void 0 : _b.length
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Fusos",
                    value: (_d = (_c = unref(data)) == null ? void 0 : _c.users_by_timezone) == null ? void 0 : _d.length
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Cidades de Vagas",
                    value: (_f = (_e = unref(data)) == null ? void 0 : _e.jobs_by_city) == null ? void 0 : _f.length
                  }, null, _parent3, _scopeId2));
                  _push3(ssrRenderComponent(_sfc_main$d, {
                    label: "Proporção",
                    value: (_g = unref(data)) == null ? void 0 : _g.remote_to_on_site_ratio
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="grid grid-cols-1 lg:grid-cols-2 gap-8"${_scopeId2}><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Usuários por Cidade</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: usersChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div><div class="h-[260px]"${_scopeId2}><h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4"${_scopeId2}>Vagas por Cidade</h3>`);
                  _push3(ssrRenderComponent(unref(Bar), {
                    data: jobsChart.value,
                    options: unref(barOptions)
                  }, null, _parent3, _scopeId2));
                  _push3(`</div></div>`);
                } else {
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Cidades",
                        value: (_i = (_h = unref(data)) == null ? void 0 : _h.users_by_city) == null ? void 0 : _i.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Fusos",
                        value: (_k = (_j = unref(data)) == null ? void 0 : _j.users_by_timezone) == null ? void 0 : _k.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Cidades de Vagas",
                        value: (_m = (_l = unref(data)) == null ? void 0 : _l.jobs_by_city) == null ? void 0 : _m.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Proporção",
                        value: (_n = unref(data)) == null ? void 0 : _n.remote_to_on_site_ratio
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Usuários por Cidade"),
                        createVNode(unref(Bar), {
                          data: usersChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Vagas por Cidade"),
                        createVNode(unref(Bar), {
                          data: jobsChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ])
                    ])
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_sfc_main$e, {
                data: unref(data),
                loading: unref(loading),
                error: unref(error)
              }, {
                default: withCtx(() => {
                  var _a, _b, _c, _d, _e, _f, _g;
                  return [
                    createVNode("div", { class: "grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8" }, [
                      createVNode(_sfc_main$d, {
                        label: "Cidades",
                        value: (_b = (_a = unref(data)) == null ? void 0 : _a.users_by_city) == null ? void 0 : _b.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Fusos",
                        value: (_d = (_c = unref(data)) == null ? void 0 : _c.users_by_timezone) == null ? void 0 : _d.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Cidades de Vagas",
                        value: (_f = (_e = unref(data)) == null ? void 0 : _e.jobs_by_city) == null ? void 0 : _f.length
                      }, null, 8, ["value"]),
                      createVNode(_sfc_main$d, {
                        label: "Proporção",
                        value: (_g = unref(data)) == null ? void 0 : _g.remote_to_on_site_ratio
                      }, null, 8, ["value"])
                    ]),
                    createVNode("div", { class: "grid grid-cols-1 lg:grid-cols-2 gap-8" }, [
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Usuários por Cidade"),
                        createVNode(unref(Bar), {
                          data: usersChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ]),
                      createVNode("div", { class: "h-[260px]" }, [
                        createVNode("h3", { class: "text-xs text-slate-400 uppercase tracking-widest mb-4" }, "Vagas por Cidade"),
                        createVNode(unref(Bar), {
                          data: jobsChart.value,
                          options: unref(barOptions)
                        }, null, 8, ["data", "options"])
                      ])
                    ])
                  ];
                }),
                _: 1
              }, 8, ["data", "loading", "error"])
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup$1 = _sfc_main$1.setup;
_sfc_main$1.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/components/analytics/GeographicWidget.vue");
  return _sfc_setup$1 ? _sfc_setup$1(props, ctx) : void 0;
};
const _sfc_main = {
  __name: "Analytics",
  __ssrInlineRender: true,
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "space-y-12" }, _attrs))}><div class="flex items-end justify-between"><div><h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Analytics</h1><p class="text-slate-400">Indicadores de matching, currículos, vagas e usuários.</p></div></div>`);
      _push(ssrRenderComponent(_sfc_main$c, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$a, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$9, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$8, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$7, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$6, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$5, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$3, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$2, null, null, _parent));
      _push(ssrRenderComponent(_sfc_main$1, null, null, _parent));
      _push(`</div>`);
    };
  }
};
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("src/pages/Analytics.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};
const routes = [
  { path: "/login", name: "Login", component: _sfc_main$p, meta: { guest: true } },
  { path: "/register", name: "Register", component: _sfc_main$o, meta: { guest: true } },
  { path: "/", name: "Dashboard", component: _sfc_main$x, meta: { requiresAuth: true } },
  { path: "/jobs", name: "Jobs", component: _sfc_main$w, meta: { requiresAuth: true } },
  { path: "/jobs/new", name: "JobAdd", component: _sfc_main$v, meta: { requiresAuth: true } },
  { path: "/matches", name: "Matches", component: _sfc_main$u, meta: { requiresAuth: true } },
  { path: "/applied", name: "Applied", component: _sfc_main$t, meta: { requiresAuth: true } },
  { path: "/analysis", name: "Analysis", component: _sfc_main$s, meta: { requiresAuth: true } },
  { path: "/analytics", name: "Analytics", component: _sfc_main, meta: { requiresAuth: true } },
  { path: "/settings", name: "Settings", component: _sfc_main$r, meta: { requiresAuth: true } },
  { path: "/settings/billing", name: "Billing", component: _sfc_main$q, meta: { requiresAuth: true } },
  { path: "/resume-editor/:id", name: "ResumeEditor", component: _sfc_main$j, meta: { requiresAuth: true } },
  { path: "/interview-prep", name: "InterviewPreparations", component: _sfc_main$h, meta: { requiresAuth: true } },
  { path: "/interview-prep/avulsa", name: "SpontaneousPrep", component: _sfc_main$g, meta: { requiresAuth: true } },
  { path: "/interview-prep/:id", name: "InterviewPrep", component: _sfc_main$i, meta: { requiresAuth: true } }
];
const router = createRouter({
  history: createWebHistory(),
  routes
});
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("vagai_token");
  const isAuth = !!token;
  if (to.meta.requiresAuth && !isAuth) {
    next({ name: "Login" });
  } else if (to.meta.guest && isAuth) {
    next({ name: "Dashboard" });
  } else if (to.meta.admin) {
    const rawUser = localStorage.getItem("vagai_user");
    const isAdmin = rawUser ? JSON.parse(rawUser).role === "admin" : false;
    next(isAdmin ? void 0 : { name: "Dashboard" });
  } else {
    next();
  }
});
export {
  router as default
};
