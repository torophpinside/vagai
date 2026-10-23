<template>
  <div class="sabatina-questions">
    <h2>Perguntas Situacionais da Sabatina</h2>
    <ul v-if="questions.length">
      <li v-for="(q, idx) in questions" :key="q.id">
        {{ idx + 1 }}. {{ q.text }}
      </li>
    </ul>
    <p v-else>Nenhuma pergunta disponível.</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const questions = ref([]);

onMounted(async () => {
  const sessionId = route.params.id;
  const res = await fetch(`/api/sabatinas/${sessionId}/questions`);
  if (res.ok) {
    const data = await res.json();
    questions.value = data.questions || [];
  }
});
</script>

<style scoped>
.sabatina-questions {
  padding: 1rem;
}
</style>
