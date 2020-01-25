
export class Page {
  pageNo ? = 1;
  pageSize ? = 10;
  totalPage?: number;
  totalCount?: number;
}

export class PageState {
  page ? = new Page();
  constructor(page?: Page) {
    if (page) {
      this.page = Object.assign(this.page, page);
    }
  }
}
