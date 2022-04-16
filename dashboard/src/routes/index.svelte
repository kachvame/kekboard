<script context="module">
  export const load = async ({ fetch }) => {
    const stats = await fetch('/stats').then((res) => res.json());

    return {
      props: {
        topScorers: stats.slice(0, 3),
        rows: stats.slice(3).map((stat) => ({
          ...stat,
          username: stat.username.split('#')[0],
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

<Card data={topScorers} />
<DataTable {headers} {rows} />
