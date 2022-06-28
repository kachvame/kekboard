<header class="header">
  <div class="header-kek">
    <img alt="kekw" src="kek.jpg" class="header-kek" />
  </div>
  <h1 class="header-title">kekboard</h1>
</header>

<style lang="scss">
  @import '@carbon/colors';

  .header {
    display: flex;
    place-items: center;
    justify-content: center;
    gap: 2vw;
    background-color: $gray-100;
    padding-top: 2vmin;
    padding-bottom: 2vmin;
  }

  .header-title {
    text-align: center;
    font-size: min(2.625rem, 10vw);
  }

  .header-kek {
    text-align: center;
  }

  .header-kek > img {
    height: min(6.25rem, 20vw);
  }
</style>
