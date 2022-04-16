<script>
  export let data;
</script>

<div class="wrapper">
  {#each data as { username, count, avatar }, index}
    <div class="card-body">
      <div class="crown-{index}" />
      <div class="avatar">
        <img class="avatar" src={avatar} alt={username} />
      </div>
      <h2 class="card-title">{username.split('#')[0]}</h2>
      <div class="card-count">{count}</div>
    </div>
  {/each}
</div>

<style>
  .wrapper {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    margin: 0 auto;
    width: 100%;
  }

  .avatar {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
  }

  .card-body {
    position: relative;
    padding: 1rem;
    height: 100%;
    width: 30%;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .card-count::before {
    content: 'Kek count: ';
    display: inline-block;
    font-weight: 500;
  }

  .card-count {
    font-size: 1.2rem;
    font-weight: 700;

    width: fit-content;
    margin: 0 auto auto;
  }

  .crown-0 {
    position: absolute;
    top: 0.2rem;
    left: 50%;
    right: 50%;
    transform: translate(-50%, -50%);
    width: 2rem;
    height: 2rem;
    background-image: url(https://svgheart.com/wp-content/uploads/2020/10/-252.png);
    background-size: 2rem;
  }
</style>
