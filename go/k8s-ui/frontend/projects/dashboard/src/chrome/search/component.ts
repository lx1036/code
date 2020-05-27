

import {Component, OnInit} from '@angular/core';
import {NgForm} from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import {SEARCH_QUERY_STATE_PARAM} from '../../common/params/params';
import {ParamsService} from '../../common/services/global/params';

@Component({
  selector: 'kd-search',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class SearchComponent implements OnInit {
  query: string;

  constructor(
    private readonly router_: Router,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly paramsService_: ParamsService,
  ) {}

  ngOnInit(): void {
    this.activatedRoute_.queryParamMap.subscribe(paramMap => {
      this.query = paramMap.get(SEARCH_QUERY_STATE_PARAM);
      this.paramsService_.setQueryParam(SEARCH_QUERY_STATE_PARAM, this.query);
    });
  }

  submit(form: NgForm): void {
    if (form.valid) {
      this.router_.navigate(['search'], {
        queryParamsHandling: 'merge',
        queryParams: {[SEARCH_QUERY_STATE_PARAM]: this.query},
      });
    }
  }
}
