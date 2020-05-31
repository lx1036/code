

import {Component, Input, OnInit} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {PolicyRule} from 'typings/backendapi';

@Component({
  selector: 'kd-policy-rule-list',
  templateUrl: './template.html',
})
export class PolicyRuleListComponent implements OnInit {
  @Input() initialized: boolean;
  @Input() rules: PolicyRule[];

  ngOnInit(): void {
    // Timeout to make sure that input variables are available.
    setTimeout(() => {
      // Filter out empty api groups.
      if (this.rules) {
        this.rules.forEach(rule => {
          if (rule.apiGroups) {
            rule.apiGroups = rule.apiGroups.filter(group => {
              return group.length > 0;
            });
          }
        });
      }
    }, 0);
  }

  getRuleColumns(): string[] {
    return ['resources', 'nonResourceURLs', 'resourceNames', 'verbs', 'apiGroups'];
  }

  getDataSource(): MatTableDataSource<PolicyRule> {
    const tableData = new MatTableDataSource<PolicyRule>();
    tableData.data = this.rules;

    return tableData;
  }
}
