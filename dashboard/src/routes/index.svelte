<script context="module">
  export const load = async ({ fetch }) => {
    const stats = await fetch('/stats').then((res) => res.json());

    return {
      props: {
        topScorers: stats.slice(0, 3),
        rows: stats.slice(3).map((stat) => ({
          ...stat,
          id: stat.username,
        })),
      },
    };
  };
</script>

<script>
  import 'carbon-components-svelte/css/g90.css';
  import { DataTable } from 'carbon-components-svelte';
  import Card from '../components/Card.svelte';
  import Header from '../components/Header.svelte';

  const headers = [
    {
      key: 'username',
      value: 'Username',
    },
    {
      key: 'count',
      value: 'Count',
    },
  ];

  export let rows;
  export let topScorers;
</script>

<svelte:head>
  <title>kekboard</title>
</svelte:head>
<Header />

<div class="container">
  <div class="wrapper">
    <div class="top-scores">
      <Card data={topScorers} />
    </div>
    <DataTable {headers} {rows} />
  </div>
</div>

<style>
  .container {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .wrapper {
    width: min(100%, 50rem);
  }

  .top-scores {
    margin-top: 2rem;
    margin-bottom: 2rem;
  }
</style>
