import { SnmpcollectorPage } from './app.po';

describe('snmpcollector App', function() {
  let page: SnmpcollectorPage;

  beforeEach(() => {
    page = new SnmpcollectorPage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
